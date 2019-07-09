package minipool

import (
    "context"
    "errors"
    "fmt"

    "github.com/docker/docker/api/types"
    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode/shared/services"
)


// Stop all running minipool containers
func stopMinipoolContainers(c *cli.Context, imageName string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        Docker: true,
    })
    if err != nil {
        return err
    }

    // Get docker containers
    containers, err := p.Docker.ContainerList(context.Background(), types.ContainerListOptions{All: true})
    if err != nil {
        return errors.New("Error retrieving docker containers: " + err.Error())
    }

    // Stop and remove minipool containers
    for _, container := range containers {
        if container.Image != imageName { continue }

        // Stop
        if err := p.Docker.ContainerStop(context.Background(), container.ID, nil); err != nil {
            return errors.New(fmt.Sprintf("Error stopping minipool container %s: " + err.Error(), container.Names[0]))
        }

        // Remove
        if err := p.Docker.ContainerRemove(context.Background(), container.ID, types.ContainerRemoveOptions{
            RemoveVolumes: true,
            RemoveLinks: true,
            Force: true,
        }); err != nil {
            return errors.New(fmt.Sprintf("Error removing minipool container %s: " + err.Error(), container.Names[0]))
        }

    }

    // Log & return
    fmt.Println("Successfully stopped all running minipool containers")
    return nil

}

