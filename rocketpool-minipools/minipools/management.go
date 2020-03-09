package minipools

import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "io/ioutil"
    "strings"
    "time"

    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/minipool"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/node"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Config
const CONTAINER_BASE_PATH string = "/.rocketpool"
const CHECK_MINIPOOLS_INTERVAL string = "1m"
var checkMinipoolsInterval, _ = time.ParseDuration(CHECK_MINIPOOLS_INTERVAL)


// Management process
type ManagementProcess struct {
    p *services.Provider
    rpPath string
    imageName string
    containerPrefix string
    rpNetwork string
}


/**
 * Start minipools management process
 */
func StartManagementProcess(p *services.Provider, rpPath string, imageName string, containerPrefix string, rpNetwork string) {

    // Initialise process
    process := &ManagementProcess{
        p: p,
        rpPath: rpPath,
        imageName: imageName,
        containerPrefix: containerPrefix,
        rpNetwork: rpNetwork,
    }

    // Start
    process.start()

}


/**
 * Start process
 */
func (p *ManagementProcess) start() {

    // Pull down minipool image
    if err := p.pullMinipoolImage(); err != nil {
        p.p.Log.Println(err)
        return
    }

    // Check minipools on interval
    go (func() {
        p.checkMinipools()
        checkMinipoolsTimer := time.NewTicker(checkMinipoolsInterval)
        for _ = range checkMinipoolsTimer.C {
            p.checkMinipools()
        }
    })()

}


/**
 * Pull down minipool image
 */
func (p *ManagementProcess) pullMinipoolImage() error {

    // Pull image
    if rc, err := p.p.Docker.ImagePull(context.Background(), p.imageName, types.ImagePullOptions{}); err != nil {
        return errors.New("Error loading minipool image: " + err.Error())
    } else {
        defer rc.Close()

        // Read docker response
        if _, err := ioutil.ReadAll(rc); err != nil {
            return errors.New("Error reading load minipool image response: " + err.Error())
        }

    }

    // Return
    return nil

}


/**
 * Check minipools
 */
func (p *ManagementProcess) checkMinipools() {

    // Data channels
    activeMinipoolAddressesChannel := make(chan []*common.Address)
    minipoolContainersChannel := make(chan []types.Container)
    errorChannel := make(chan error)

    // Wait for node to sync
    eth.WaitSync(p.p.Client, true, false)

    // Get active minipool addresses
    go (func() {

        // Get minipool addresses
        nodeAccount, _ := p.p.AM.GetNodeAccount()
        minipoolAddresses, err := node.GetMinipoolAddresses(nodeAccount.Address, p.p.CM)
        if err != nil {
            errorChannel <- err
            return
        }
        minipoolCount := len(minipoolAddresses)

        // Get minipool statuses
        statusChannels := make([]chan uint8, minipoolCount)
        statusErrorChannel := make(chan error)
        for mi := 0; mi < minipoolCount; mi++ {
            statusChannels[mi] = make(chan uint8)
            go (func(mi int) {
                if status, err := minipool.GetStatusCode(p.p.CM, minipoolAddresses[mi]); err != nil {
                    statusErrorChannel <- err
                } else {
                    statusChannels[mi] <- status
                }
            })(mi)
        }

        // Receive minipool statuses & filter active minipools
        activeMinipoolAddresses := []*common.Address{}
        for mi := 0; mi < minipoolCount; mi++ {
            select {
                case status := <-statusChannels[mi]:
                    if status == minipool.PRELAUNCH || status == minipool.STAKING { activeMinipoolAddresses = append(activeMinipoolAddresses, minipoolAddresses[mi]) }
                case err := <-statusErrorChannel:
                    errorChannel <- err
                    return
            }
        }

        // Send active minipool addresses
        activeMinipoolAddressesChannel <- activeMinipoolAddresses

    })()

    // Get active minipool containers
    go (func() {

        // Get docker containers
        containers, err := p.p.Docker.ContainerList(context.Background(), types.ContainerListOptions{All: true})
        if err != nil {
            errorChannel <- errors.New("Error retrieving docker containers: " + err.Error())
            return
        }

        // Filter by minipool container image name
        minipoolContainers := []types.Container{}
        for _, container := range containers {
            if container.Image == p.imageName { minipoolContainers = append(minipoolContainers, container) }
        }

        // Send minipool containers
        minipoolContainersChannel <- minipoolContainers

    })()

    // Receive minipool data
    var activeMinipoolAddresses []*common.Address
    var minipoolContainers []types.Container
    for received := 0; received < 2; {
        select {
            case activeMinipoolAddresses = <-activeMinipoolAddressesChannel:
                received++
            case minipoolContainers = <-minipoolContainersChannel:
                received++
            case err := <-errorChannel:
                p.p.Log.Println(err)
                return
        }
    }

    // Run minipool containers
    for _, minipoolAddress := range activeMinipoolAddresses {
        go p.runMinipoolContainer(minipoolAddress, minipoolContainers)
    }

    // Check inactive minipool containers
    for _, container := range minipoolContainers {
        if container.State == "exited" {
            go p.checkInactiveMinipoolContainer(container, activeMinipoolAddresses)
        }
    }

}


/**
 * Run minipool container
 */
func (p *ManagementProcess) runMinipoolContainer(minipoolAddress *common.Address, minipoolContainers []types.Container) {

    // Get name for minipool container
    containerName := p.containerPrefix + minipoolAddress.Hex()

    // Get existing minipool container ID
    var containerId string
    for _, container := range minipoolContainers {
        if "/" + containerName == container.Names[0] {
            containerId = container.ID
            break
        }
    }

    // Create minipool container if it doesn't exist
    if containerId == "" {

        // Log
        p.p.Log.Println(fmt.Sprintf("Creating minipool container %s...", containerName))

        // Create container
        if response, err := p.p.Docker.ContainerCreate(context.Background(), &container.Config{
            Image: p.imageName,
            Cmd: []string{minipoolAddress.Hex()},
        }, &container.HostConfig{
            Binds: []string{p.rpPath + ":" + CONTAINER_BASE_PATH},
            NetworkMode: container.NetworkMode(p.rpNetwork),
            RestartPolicy: container.RestartPolicy{Name: "on-failure"},
        }, nil, containerName); err != nil {
            p.p.Log.Println(errors.New(fmt.Sprintf("Error creating minipool container %s: " + err.Error(), containerName)))
            return
        } else {
            p.p.Log.Println(fmt.Sprintf("Created minipool container %s successfully", containerName))
            containerId = response.ID
        }

    }

    // Start minipool container if not running
    if container, err := p.p.Docker.ContainerInspect(context.Background(), containerId); err != nil {
        p.p.Log.Println(errors.New(fmt.Sprintf("Error inspecting minipool container %s: " + err.Error(), containerName)))
        return
    } else if !container.State.Running {

        // Log
        p.p.Log.Println(fmt.Sprintf("Starting minipool container %s...", containerName))

        // Start container
        if err := p.p.Docker.ContainerStart(context.Background(), containerId, types.ContainerStartOptions{}); err != nil {
            p.p.Log.Println(errors.New(fmt.Sprintf("Error starting minipool container %s: " + err.Error(), containerName)))
            return
        } else {
            p.p.Log.Println(fmt.Sprintf("Started minipool container %s successfully", containerName))
        }

    }

}


/**
 * Check inactive minipool container
 */
func (p *ManagementProcess) checkInactiveMinipoolContainer(container types.Container, activeMinipoolAddresses []*common.Address) {

    // Get minipool container name
    containerName := container.Names[0]

    // Get minipool container address from command args
    args := strings.Split(container.Command, " ")
    addressStr := args[len(args) - 1]

    // Check and parse minipool container address
    if !common.IsHexAddress(addressStr) {
        p.p.Log.Println(errors.New(fmt.Sprintf("Could not get minipool address for container %s", containerName)))
        return
    }
    address := common.HexToAddress(addressStr)

    // Check if minipool is active
    active := false
    for _, activeMinipoolAddress := range activeMinipoolAddresses {
        if bytes.Equal(address.Bytes(), activeMinipoolAddress.Bytes()) {
            active = true
            break
        }
    }
    if active { return }

    // Remove container
    p.p.Log.Println(fmt.Sprintf("Removing minipool container %s...", containerName))
    if err := p.p.Docker.ContainerRemove(context.Background(), container.ID, types.ContainerRemoveOptions{}); err != nil {
        p.p.Log.Println(errors.New(fmt.Sprintf("Error removing minipool container %s: " + err.Error(), containerName)))
    } else {
        p.p.Log.Println(fmt.Sprintf("Removed minipool container %s successfully", containerName))
    }

}

