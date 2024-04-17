package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	dt "github.com/docker/docker/api/types"
	dtc "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dti "github.com/docker/docker/api/types/image"
)

// Get the current Docker image used by the given container
func (c *Client) GetDockerImage(containerName string) (string, error) {
	ci, err := inspectContainer(c, containerName)
	if err != nil {
		return "", err
	}
	return ci.Config.Image, nil
}

// Check if a container with the provided name exists
func (c *Client) CheckIfContainerExists(containerName string) (bool, error) {
	d, err := c.GetDocker()
	if err != nil {
		return false, err
	}
	cl, err := d.ContainerList(context.Background(), dtc.ListOptions{
		All: true, Filters: filters.NewArgs(
			filters.Arg("name", containerName),
		),
	})
	if err != nil {
		return false, fmt.Errorf("error getting container list: %w", err)
	}
	return len(cl) > 0, nil
}

// Get the current Docker image used by the given container
func (c *Client) GetDockerStatus(containerName string) (string, error) {
	ci, err := inspectContainer(c, containerName)
	if err != nil {
		return "", err
	}
	return ci.State.Status, nil
}

// Get the time that the given container shut down
func (c *Client) GetDockerContainerShutdownTime(containerName string) (time.Time, error) {
	ci, err := inspectContainer(c, containerName)
	if err != nil {
		return time.Time{}, err
	}

	// Parse the time
	finishTime, err := time.Parse(time.RFC3339, strings.TrimSpace(ci.State.FinishedAt))
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing container [%s] exit time [%s]: %w", containerName, ci.State.FinishedAt, err)
	}
	return finishTime, nil
}

// Shut down a container
func (c *Client) StopContainer(containerName string) error {
	d, err := c.GetDocker()
	if err != nil {
		return err
	}
	return d.ContainerStop(context.Background(), containerName, dtc.StopOptions{})
}

// Start a container
func (c *Client) StartContainer(containerName string) error {
	d, err := c.GetDocker()
	if err != nil {
		return err
	}
	return d.ContainerStart(context.Background(), containerName, dtc.StartOptions{})
}

// Restart a container
func (c *Client) RestartContainer(containerName string) error {
	d, err := c.GetDocker()
	if err != nil {
		return err
	}
	return d.ContainerRestart(context.Background(), containerName, dtc.StopOptions{})
}

// Deletes a container
func (c *Client) RemoveContainer(containerName string) error {
	d, err := c.GetDocker()
	if err != nil {
		return err
	}
	return d.ContainerRemove(context.Background(), containerName, dtc.RemoveOptions{})
}

// Deletes a volume
func (c *Client) DeleteVolume(volumeName string) error {
	d, err := c.GetDocker()
	if err != nil {
		return err
	}
	return d.VolumeRemove(context.Background(), volumeName, false)
}

// Deletes an image
func (c *Client) DeleteImage(imageName string) error {
	d, err := c.GetDocker()
	if err != nil {
		return err
	}
	// TODO: handle the response here
	_, err = d.ImageRemove(context.Background(), imageName, dti.RemoveOptions{})
	return err
}

// Gets the absolute file path of the client volume
func (c *Client) GetClientVolumeSource(containerName string, volumeTarget string) (string, error) {
	ci, err := inspectContainer(c, containerName)
	if err != nil {
		return "", err
	}

	// Find the mount with the provided destination
	for _, mount := range ci.Mounts {
		if mount.Destination == volumeTarget {
			return mount.Source, nil
		}
	}
	return "", fmt.Errorf("container [%s] doesn't have a volume with [%s] as a destination", containerName, volumeTarget)
}

// Gets the name of the client volume
func (c *Client) GetClientVolumeName(containerName, volumeTarget string) (string, error) {
	ci, err := inspectContainer(c, containerName)
	if err != nil {
		return "", err
	}

	// Find the mount with the provided destination
	for _, mount := range ci.Mounts {
		if mount.Destination == volumeTarget {
			return mount.Name, nil
		}
	}
	return "", fmt.Errorf("container [%s] doesn't have a volume with [%s] as a destination", containerName, volumeTarget)
}

// Gets the disk usage of the given volume
func (c *Client) GetVolumeSize(volumeName string) (int64, error) {
	d, err := c.GetDocker()
	if err != nil {
		return 0, err
	}

	du, err := d.DiskUsage(context.Background(), dt.DiskUsageOptions{})
	if err != nil {
		return 0, fmt.Errorf("error getting disk usage: %w", err)
	}
	for _, volume := range du.Volumes {
		if volume.Name == volumeName {
			return volume.UsageData.Size, nil
		}
	}
	return 0, fmt.Errorf("couldn't find a volume named [%s]", volumeName)
}

// Derived from https://github.com/docker/cli/blob/ced099660009713e0e845eeb754e6050dbaa45d0/cli/command/system/prune.go#L71
func (c *Client) DockerSystemPrune(deleteAllImages bool) error {
	d, err := c.GetDocker()
	if err != nil {
		return err
	}

	// Prune containers
	emptyFilters := filters.NewArgs()
	_, err = d.ContainersPrune(context.Background(), emptyFilters)
	if err != nil {
		return fmt.Errorf("error pruning containers: %w", err)
	}

	// Prune images
	_, err = d.NetworksPrune(context.Background(), emptyFilters)
	if err != nil {
		return fmt.Errorf("error pruning networks: %w", err)
	}

	// Prune images - this is a little silly but they don't include an "all" flag in the API so we have to reimplement the CLI
	// See https://github.com/docker/cli/blob/38fcd1ca63d8cb44299d0c6e0911f861211cacde/cli/command/image/prune.go#L63
	pruneFilters := filters.NewArgs(filters.KeyValuePair{
		Key:   "dangling",
		Value: strconv.FormatBool(!deleteAllImages),
	})
	_, err = d.ImagesPrune(context.Background(), pruneFilters)
	if err != nil {
		return fmt.Errorf("error pruning images: %w", err)
	}

	return nil
}

// Returns all Docker images on the system
func (c *Client) GetAllDockerImages() ([]DockerImage, error) {
	d, err := c.GetDocker()
	if err != nil {
		return nil, err
	}

	imageList, err := d.ImageList(context.Background(), dti.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("error getting image details: %w", err)
	}

	images := make([]DockerImage, len(imageList))
	for i, image := range imageList {
		images[i].ID = image.ID
		images[i].RepositoryTag = image.RepoTags[0]
	}

	return images, nil
}

// Inspect a Docker container
func inspectContainer(c *Client, container string) (dt.ContainerJSON, error) {
	d, err := c.GetDocker()
	if err != nil {
		return dt.ContainerJSON{}, err
	}
	ci, err := d.ContainerInspect(context.Background(), container)
	if err != nil {
		return dt.ContainerJSON{}, fmt.Errorf("error inspecting container [%s]: %w", container, err)
	}
	return ci, nil
}
