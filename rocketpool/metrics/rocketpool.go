package metrics

import (
    "bytes"
    "errors"
    "fmt"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/minipool"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/settings"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// RP metrics process
type RocketPoolMetricsProcess struct {
    p                       *services.Provider
    stakingDurationEnabled  map[string]prometheus.Gauge
    networkEthCapacity      map[string]prometheus.Gauge
    networkEthAssigned      map[string]prometheus.Gauge
    networkUtilisation      map[string]prometheus.Gauge
    rplRatio                map[string]prometheus.Gauge
    queueBalance            map[string]prometheus.Gauge
    totalMinipools          prometheus.Gauge
    statusMinipools         map[uint8]prometheus.Gauge
    totalNodes              prometheus.Gauge
    activeNodes             prometheus.Gauge
}


// Staking duration metrics
type StakingDurationMetrics struct {
    DurationId string
    NetworkEthCapacity float64
    NetworkEthAssigned float64
    NetworkUtilisation float64
    RplRatio float64
    QueueBalance float64
}


// Start RP metrics process
func StartRocketPoolMetricsProcess(p *services.Provider) {

    // Initialise process / register metrics
    process := &RocketPoolMetricsProcess{
        p:                      p,
        stakingDurationEnabled: make(map[string]prometheus.Gauge),
        networkEthCapacity:     make(map[string]prometheus.Gauge),
        networkEthAssigned:     make(map[string]prometheus.Gauge),
        networkUtilisation:     make(map[string]prometheus.Gauge),
        rplRatio:               make(map[string]prometheus.Gauge),
        queueBalance:           make(map[string]prometheus.Gauge),
        totalMinipools:         promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:  "smartnode",
            Subsystem:  "rocketpool",
            Name:       "minipool_total_count",
            Help:       "The total number of minipools in Rocket Pool",
        }),
        statusMinipools:        make(map[uint8]prometheus.Gauge),
        totalNodes:             promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:  "smartnode",
            Subsystem:  "rocketpool",
            Name:       "node_total_count",
            Help:       "The total number of nodes in Rocket Pool",
        }),
        activeNodes:            promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:  "smartnode",
            Subsystem:  "rocketpool",
            Name:       "node_active_count",
            Help:       "The number of active nodes in Rocket Pool",
        }),
    }
    for status := uint8(0); status < uint8(8); status++ {
        statusType := minipool.GetStatusType(status)
        process.statusMinipools[status] = promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:  "smartnode",
            Subsystem:  "rocketpool",
            Name:       fmt.Sprintf("minipool_%s_count", statusType),
            Help:       fmt.Sprintf("The number of '%s' minipools in Rocket Pool", statusType),
        })
    }

    // Start
    process.start()

}


// Start process
func (p *RocketPoolMetricsProcess) start() {

    // Update metrics on interval
    go (func() {
        p.update()
        updateMetricsTimer := time.NewTicker(updateMetricsInterval)
        for _ = range updateMetricsTimer.C {
            p.update()
        }
    })()

}


// Update metrics
func (p *RocketPoolMetricsProcess) update() {
    go p.updateStakingDurationMetrics()
    go p.updateMinipoolMetrics()
    go p.updateNodeMetrics()
}


// Update staking duration metrics
func (p *RocketPoolMetricsProcess) updateStakingDurationMetrics() {

    // Get staking durations
    stakingDurations, err := settings.GetMinipoolStakingDurations(p.p.CM)
    if err != nil {
        p.p.Log.Println(err)
        return
    }
    stakingDurationCount := len(stakingDurations)

    // Get staking duration metrics
    stakingDurationMetricsChannels := make([]chan *StakingDurationMetrics, stakingDurationCount)
    errorChannel := make(chan error)
    for di := 0; di < stakingDurationCount; di++ {
        stakingDurationMetricsChannels[di] = make(chan *StakingDurationMetrics)
        go (func(di int) {
            if metrics, err := getStakingDurationMetrics(p.p.CM, stakingDurations[di].Id); err != nil {
                errorChannel <- err
            } else {
                stakingDurationMetricsChannels[di] <- metrics
            }
        })(di)
    }

    // Receive staking duration metrics
    stakingDurationMetrics := make([]*StakingDurationMetrics, stakingDurationCount)
    for di := 0; di < stakingDurationCount; di++ {
        select {
            case metrics := <-stakingDurationMetricsChannels[di]:
                stakingDurationMetrics[di] = metrics
            case err := <-errorChannel:
                p.p.Log.Println(err)
                return
        }
    }

    // Create & update metrics
    for di := 0; di < stakingDurationCount; di++ {
        duration := stakingDurations[di]
        metrics := stakingDurationMetrics[di]

        // Staking duration enabled
        if _, ok := p.stakingDurationEnabled[duration.Id]; !ok {
            p.stakingDurationEnabled[duration.Id] = promauto.NewGauge(prometheus.GaugeOpts{
                Namespace:  "smartnode",
                Subsystem:  "rocketpool",
                Name:       fmt.Sprintf("staking_enabled_%s", duration.Id),
                Help:       fmt.Sprintf("Whether the '%s' staking duration is enabled", duration.Id),
            })
        }
        if duration.Enabled {
            p.stakingDurationEnabled[duration.Id].Set(1)
        } else {
            p.stakingDurationEnabled[duration.Id].Set(0)
        }

        // Network ETH capacity
        if _, ok := p.networkEthCapacity[duration.Id]; !ok {
            p.networkEthCapacity[duration.Id] = promauto.NewGauge(prometheus.GaugeOpts{
                Namespace:  "smartnode",
                Subsystem:  "rocketpool",
                Name:       fmt.Sprintf("network_eth_capacity_%s", duration.Id),
                Help:       fmt.Sprintf("The network ETH capacity for the '%s' queue", duration.Id),
            })
        }
        p.networkEthCapacity[duration.Id].Set(metrics.NetworkEthCapacity)

        // Network ETH assigned
        if _, ok := p.networkEthAssigned[duration.Id]; !ok {
            p.networkEthAssigned[duration.Id] = promauto.NewGauge(prometheus.GaugeOpts{
                Namespace:  "smartnode",
                Subsystem:  "rocketpool",
                Name:       fmt.Sprintf("network_eth_assigned_%s", duration.Id),
                Help:       fmt.Sprintf("The network ETH assigned for the '%s' queue", duration.Id),
            })
        }
        p.networkEthAssigned[duration.Id].Set(metrics.NetworkEthAssigned)

        // Network utilisation
        if _, ok := p.networkUtilisation[duration.Id]; !ok {
            p.networkUtilisation[duration.Id] = promauto.NewGauge(prometheus.GaugeOpts{
                Namespace:  "smartnode",
                Subsystem:  "rocketpool",
                Name:       fmt.Sprintf("network_utilisation_%s", duration.Id),
                Help:       fmt.Sprintf("The network utilisation for the '%s' queue", duration.Id),
            })
        }
        p.networkUtilisation[duration.Id].Set(metrics.NetworkUtilisation)

        // RPL ratio
        if _, ok := p.rplRatio[duration.Id]; !ok {
            p.rplRatio[duration.Id] = promauto.NewGauge(prometheus.GaugeOpts{
                Namespace:  "smartnode",
                Subsystem:  "rocketpool",
                Name:       fmt.Sprintf("rpl_ratio_%s", duration.Id),
                Help:       fmt.Sprintf("The RPL:ETH ratio for the '%s' queue", duration.Id),
            })
        }
        p.rplRatio[duration.Id].Set(metrics.RplRatio)

        // Queue balance
        if _, ok := p.queueBalance[duration.Id]; !ok {
            p.queueBalance[duration.Id] = promauto.NewGauge(prometheus.GaugeOpts{
                Namespace:  "smartnode",
                Subsystem:  "rocketpool",
                Name:       fmt.Sprintf("queue_balance_%s", duration.Id),
                Help:       fmt.Sprintf("The current ETH balance of the '%s' deposit queue", duration.Id),
            })
        }
        p.queueBalance[duration.Id].Set(metrics.QueueBalance)

    }

}


// Get staking duration metrics
func getStakingDurationMetrics(cm *rocketpool.ContractManager, durationId string) (*StakingDurationMetrics, error) {

    // Data channels
    networkEthCapacityChannel := make(chan float64)
    networkEthAssignedChannel := make(chan float64)
    networkUtilisationChannel := make(chan float64)
    rplRatioChannel := make(chan float64)
    queueBalanceChannel := make(chan float64)
    errorChannel := make(chan error)

    // Get network ETH capacity
    go (func() {
        networkEthCapacity := new(*big.Int)
        if err := cm.Contracts["rocketPool"].Call(nil, networkEthCapacity, "getTotalEther", "capacity", durationId); err != nil {
            errorChannel <- errors.New("Error retrieving network ETH capacity: " + err.Error())
        } else {
            networkEthCapacityChannel <- eth.WeiToEth(*networkEthCapacity)
        }
    })()

    // Get network ETH assigned
    go (func() {
        networkEthAssigned := new(*big.Int)
        if err := cm.Contracts["rocketPool"].Call(nil, networkEthAssigned, "getTotalEther", "assigned", durationId); err != nil {
            errorChannel <- errors.New("Error retrieving network ETH assigned: " + err.Error())
        } else {
            networkEthAssignedChannel <- eth.WeiToEth(*networkEthAssigned)
        }
    })()

    // Get network utilisation
    go (func() {
        networkUtilisation := new(*big.Int)
        if err := cm.Contracts["rocketPool"].Call(nil, networkUtilisation, "getNetworkUtilisation", durationId); err != nil {
            errorChannel <- errors.New("Error retrieving network utilisation: " + err.Error())
        } else {
            networkUtilisationChannel <- eth.WeiToEth(*networkUtilisation)
        }
    })()

    // Get RPL ratio
    go (func() {
        rplRatioWei := new(*big.Int)
        if err := cm.Contracts["rocketNodeAPI"].Call(nil, rplRatioWei, "getRPLRatio", durationId); err != nil {
            errorChannel <- errors.New("Error retrieving required RPL amount: " + err.Error())
        } else {
            rplRatioChannel <- eth.WeiToEth(*rplRatioWei)
        }
    })()

    // Get queue balance
    go (func() {
        queueBalanceWei := new(*big.Int)
        if err := cm.Contracts["rocketDepositQueue"].Call(nil, queueBalanceWei, "getBalance", durationId); err != nil {
            errorChannel <- errors.New("Error retrieving deposit queue balance: " + err.Error())
        } else {
            queueBalanceChannel <- eth.WeiToEth(*queueBalanceWei)
        }
    })()

    // Receive data & return
    metrics := &StakingDurationMetrics{DurationId: durationId}
    for received := 0; received < 5; {
        select {
            case metrics.NetworkEthCapacity = <-networkEthCapacityChannel:
                received++
            case metrics.NetworkEthAssigned = <-networkEthAssignedChannel:
                received++
            case metrics.NetworkUtilisation = <-networkUtilisationChannel:
                received++
            case metrics.RplRatio = <-rplRatioChannel:
                received++
            case metrics.QueueBalance = <-queueBalanceChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }
    return metrics, nil

}


// Update minipool metrics
func (p *RocketPoolMetricsProcess) updateMinipoolMetrics() {

    // Get minipool count
    minipoolCountV := new(*big.Int)
    if err := p.p.CM.Contracts["rocketPool"].Call(nil, minipoolCountV, "getPoolsCount"); err != nil {
        p.p.Log.Println(errors.New("Error retrieving minipool count: " + err.Error()))
        return
    }
    minipoolCount := (*minipoolCountV).Int64()

    // Data channels
    addressChannels := make([]chan *common.Address, minipoolCount)
    statusChannels := make([]chan uint8, minipoolCount)
    errorChannel := make(chan error)

    // Get minipool addresses
    for mi := int64(0); mi < minipoolCount; mi++ {
        addressChannels[mi] = make(chan *common.Address)
        go (func(mi int64) {
            minipoolAddress := new(common.Address)
            if err := p.p.CM.Contracts["rocketPool"].Call(nil, minipoolAddress, "getPoolAt", big.NewInt(mi)); err != nil {
                errorChannel <- errors.New("Error retrieving minipool address: " + err.Error())
            } else {
                addressChannels[mi] <- minipoolAddress
            }
        })(mi)
    }

    // Receive minipool addresses
    minipoolAddresses := make([]*common.Address, minipoolCount)
    for mi := int64(0); mi < minipoolCount; mi++ {
        select {
            case address := <-addressChannels[mi]:
                minipoolAddresses[mi] = address
            case err := <-errorChannel:
                p.p.Log.Println(err)
                return
        }
    }

    // Get minipool statuses
    for mi := int64(0); mi < minipoolCount; mi++ {
        statusChannels[mi] = make(chan uint8)
        go (func(mi int64) {

            // Initialise minipool contract
            minipoolContract, err := p.p.CM.NewContract(minipoolAddresses[mi], "rocketMinipool")
            if err != nil {
                errorChannel <- errors.New("Error initialising minipool contract: " + err.Error())
                return
            }

            // Get status
            status := new(uint8)
            if err := minipoolContract.Call(nil, status, "getStatus"); err != nil {
                errorChannel <- errors.New("Error retrieving minipool status: " + err.Error())
            } else {
                statusChannels[mi] <- *status
            }

        })(mi)
    }

    // Receive minipool statuses
    statusCounts := make(map[uint8]uint64)
    for mi := int64(0); mi < minipoolCount; mi++ {
        select {
            case status := <-statusChannels[mi]:
                if _, ok := statusCounts[status]; !ok { statusCounts[status] = 0 }
                statusCounts[status]++
            case err := <-errorChannel:
                p.p.Log.Println(err)
                return
        }
    }

    // Update minipool metrics
    p.totalMinipools.Set(float64(minipoolCount))
    for status, count := range statusCounts {
        p.statusMinipools[status].Set(float64(count))
    }

}


// Update node metrics
func (p *RocketPoolMetricsProcess) updateNodeMetrics() {

    // Get node list key
    nodeListKey := eth.KeccakBytes(bytes.Join([][]byte{[]byte("nodes"), []byte("list")}, []byte{}))

    // Get node count
    nodeCountV := new(*big.Int)
    if err := p.p.CM.Contracts["utilAddressSetStorage"].Call(nil, nodeCountV, "getCount", nodeListKey); err != nil {
        p.p.Log.Println(errors.New("Error retrieving node count: " + err.Error()))
        return
    }
    nodeCount := (*nodeCountV).Int64()

    // Data channels
    addressChannels := make([]chan *common.Address, nodeCount)
    activeChannels := make([]chan bool, nodeCount)
    errorChannel := make(chan error)

    // Get node addresses
    for ni := int64(0); ni < nodeCount; ni++ {
        addressChannels[ni] = make(chan *common.Address)
        go (func(ni int64) {
            nodeAddress := new(common.Address)
            if err := p.p.CM.Contracts["utilAddressSetStorage"].Call(nil, nodeAddress, "getItem", nodeListKey, big.NewInt(ni)); err != nil {
                errorChannel <- errors.New("Error retrieving node address: " + err.Error())
            } else {
                addressChannels[ni] <- nodeAddress
            }
        })(ni)
    }

    // Receive node addresses
    nodeAddresses := make([]*common.Address, nodeCount)
    for ni := int64(0); ni < nodeCount; ni++ {
        select {
            case address := <-addressChannels[ni]:
                nodeAddresses[ni] = address
            case err := <-errorChannel:
                p.p.Log.Println(err)
                return
        }
    }

    // Get node statuses
    for ni := int64(0); ni < nodeCount; ni++ {
        activeChannels[ni] = make(chan bool)
        go (func(ni int64) {
            nodeActiveKey := eth.KeccakBytes(bytes.Join([][]byte{[]byte("node.active"), nodeAddresses[ni].Bytes()}, []byte{}))
            if active, err := p.p.CM.RocketStorage.GetBool(nil, nodeActiveKey); err != nil {
                errorChannel <- errors.New("Error retrieving node active status: " + err.Error())
            } else {
                activeChannels[ni] <- active
            }
        })(ni)
    }

    // Receive node status
    activeNodes := 0
    for ni := int64(0); ni < nodeCount; ni++ {
        select {
            case active := <-activeChannels[ni]:
                if active { activeNodes++ }
            case err := <-errorChannel:
                p.p.Log.Println(err)
                return
        }
    }

    // Update node metrics
    p.totalNodes.Set(float64(nodeCount))
    p.activeNodes.Set(float64(activeNodes))

}

