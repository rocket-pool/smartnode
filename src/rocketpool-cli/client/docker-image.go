package client

import "fmt"

type DockerImage struct {
	ID            string
	RepositoryTag string
}

func (img *DockerImage) String() string {
	return fmt.Sprintf("%s (%s)", img.RepositoryTag, img.ID)
}
