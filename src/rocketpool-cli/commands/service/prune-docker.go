package service

import (
	"fmt"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/urfave/cli/v2"
)

func pruneDocker(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// NOTE: we deliberately avoid using `docker system prune -a` and delete all
	//   images manually so that we can preserve the current smartnode-stack
	//   images, _unless_ the user specified --all option
	deleteAllImages := c.Bool(utils.YesFlag.Name)
	if !deleteAllImages {
		ourImages, err := rp.GetComposeImages(getComposeFiles(c))
		if err != nil {
			return fmt.Errorf("error getting compose images: %w", err)
		}

		ourImagesMap := make(map[string]struct{})
		for _, image := range ourImages {
			ourImagesMap[image] = struct{}{}
		}

		allImages, err := rp.GetAllDockerImages()
		if err != nil {
			return fmt.Errorf("error getting all docker images: %w", err)
		}

		fmt.Println("Deleting images not used by the Smart Node...")
		for _, image := range allImages {
			if _, ok := ourImagesMap[image.RepositoryTag]; !ok {
				fmt.Printf("Deleting %s...\n", image.String())
				err = rp.DeleteImage(image.ID)
				if err != nil {
					// safe to ignore and print to user, since it may just be an image referenced by a running container that is managed outside of the Smart Node's compose stack
					fmt.Printf("Error deleting image %s: %s\n", image.String(), err.Error())
				}
				continue
			}

			fmt.Printf("Skipping image used by the Smart Node stack: %s\n", image.String())
		}
	}

	// now we can run docker system prune (potentially without --all) to remove
	// all stopped containers and networks:
	fmt.Println("Pruning Docker system...")
	err := rp.DockerSystemPrune(deleteAllImages)
	if err != nil {
		return fmt.Errorf("error pruning Docker system: %w", err)
	}

	return nil
}
