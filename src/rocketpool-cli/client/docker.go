package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	dt "github.com/docker/docker/api/types"
	dtc "github.com/docker/docker/api/types/container"
)

// Get the current Docker image used by the given container
func (c *Client) GetDockerImage(containerName string) (string, error) {
	ci, err := inspectContainer(c, containerName)
	if err != nil {
		return "", err
	}
	return ci.Config.Image, nil
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
	return d.ContainerStart(context.Background(), containerName, dt.ContainerStartOptions{})
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
	return d.ContainerRemove(context.Background(), containerName, dt.ContainerRemoveOptions{})
}

// Deletes a container
func (c *Client) DeleteVolume(volumeName string) error {
	d, err := c.GetDocker()
	if err != nil {
		return err
	}
	return d.VolumeRemove(context.Background(), volumeName, false)
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
