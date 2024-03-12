package client

import (
	"fmt"
	"strings"
	"time"
)

// Get the current Docker image used by the given container
func (c *Client) GetDockerImage(container string) (string, error) {

	cmd := fmt.Sprintf("docker container inspect --format={{.Config.Image}} %s", container)
	image, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(image)), nil

}

// Get the current Docker image used by the given container
func (c *Client) GetDockerStatus(container string) (string, error) {

	cmd := fmt.Sprintf("docker container inspect --format={{.State.Status}} %s", container)
	status, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(status)), nil

}

// Get the time that the given container shut down
func (c *Client) GetDockerContainerShutdownTime(container string) (time.Time, error) {

	cmd := fmt.Sprintf("docker container inspect --format={{.State.FinishedAt}} %s", container)
	finishTimeBytes, err := c.readOutput(cmd)
	if err != nil {
		return time.Time{}, err
	}

	finishTime, err := time.Parse(time.RFC3339, strings.TrimSpace(string(finishTimeBytes)))
	if err != nil {
		return time.Time{}, fmt.Errorf("Error parsing validator container exit time [%s]: %w", string(finishTimeBytes), err)
	}

	return finishTime, nil

}

// Shut down a container
func (c *Client) StopContainer(container string) (string, error) {

	cmd := fmt.Sprintf("docker stop %s", container)
	output, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil

}

// Start a container
func (c *Client) StartContainer(container string) (string, error) {

	cmd := fmt.Sprintf("docker start %s", container)
	output, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil

}

// Restart a container
func (c *Client) RestartContainer(container string) (string, error) {

	cmd := fmt.Sprintf("docker restart %s", container)
	output, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil

}

// Deletes a container
func (c *Client) RemoveContainer(container string) (string, error) {

	cmd := fmt.Sprintf("docker rm %s", container)
	output, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil

}

// Deletes a container
func (c *Client) DeleteVolume(volume string) (string, error) {

	cmd := fmt.Sprintf("docker volume rm %s", volume)
	output, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil

}

// Gets the absolute file path of the client volume
func (c *Client) GetClientVolumeSource(container string, volumeTarget string) (string, error) {

	cmd := fmt.Sprintf("docker container inspect --format='{{range .Mounts}}{{if eq \"%s\" .Destination}}{{.Source}}{{end}}{{end}}' %s", volumeTarget, container)
	output, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// Gets the name of the client volume
func (c *Client) GetClientVolumeName(container string, volumeTarget string) (string, error) {

	cmd := fmt.Sprintf("docker container inspect --format='{{range .Mounts}}{{if eq \"%s\" .Destination}}{{.Name}}{{end}}{{end}}' %s", volumeTarget, container)
	output, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// Gets the disk usage of the given volume
func (c *Client) GetVolumeSize(volumeName string) (string, error) {

	cmd := fmt.Sprintf("docker system df -v --format='{{range .Volumes}}{{if eq \"%s\" .Name}}{{.Size}}{{end}}{{end}}'", volumeName)
	output, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
