package main

import (
	"fmt"
	"go/build"
	"os"

	"github.com/princjef/gomarkdoc"
	"github.com/princjef/gomarkdoc/lang"
	"github.com/princjef/gomarkdoc/logger"
)

const repo string = "https://github.com/rocket-pool/rocketpool-go"
const branch string = "release"

func main() {

	// gomarkdoc's logger
	log := logger.New(logger.DebugLevel)

	// Make a new doc renderer
	renderer, err := gomarkdoc.NewRenderer()
	if err != nil {
		fmt.Printf("Error creating renderer: %s\n", err.Error())
		os.Exit(1)
	}

	// Get the working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting working directory: %s\n", err.Error())
		os.Exit(1)
	}

	// These are all of the packages to generate the source for
	packages := map[string]string{
		"auction":              "%s/../auction",
		"contracts":            "%s/../contracts",
		"dao":                  "%s/../dao",
		"dao-protocol":         "%s/../dao/protocol",
		"dao-trustednode":      "%s/../dao/trustednode",
		"deposit":              "%s/../deposit",
		"minipool":             "%s/../minipool",
		"network":              "%s/../network",
		"node":                 "%s/../node",
		"rewards":              "%s/../rewards",
		"rocketpool":           "%s/../rocketpool",
		"settings-protocol":    "%s/../settings/protocol",
		"settings-trustednode": "%s/../settings/trustednode",
		"storage":              "%s/../storage",
		"tokens":               "%s/../tokens",
		"types":                "%s/../types",
		"utils":                "%s/../utils",
		"utils-eth":            "%s/../utils/eth",
		"utils-strings":        "%s/../utils/strings",
	}

	// Build the documentation file for each package
	for filename, path := range packages {

		// Load the source dir
		builder, err := build.ImportDir(fmt.Sprintf(path, wd), build.ImportComment)
		if err != nil {
			fmt.Printf("Error loading package builder for %s: %s\n", filename, err.Error())
			os.Exit(1)
		}

		// Create a package from the source
		pkg, err := lang.NewPackageFromBuild(log, builder, lang.PackageWithRepositoryOverrides(&lang.Repo{
			Remote:        repo,
			DefaultBranch: branch,
		}))
		if err != nil {
			fmt.Printf("Error creating package %s: %s\n", filename, err.Error())
			os.Exit(1)
		}

		// Render the documentation for the package
		packageContents, err := renderer.Package(pkg)
		if err != nil {
			fmt.Printf("Error exporting package %s: %s\n", filename, err.Error())
			os.Exit(1)
		}

		// Write the docs out to the appropriate file
		err = os.WriteFile(fmt.Sprintf("%s/%s.md", wd, filename), []byte(packageContents), 0644)
		if err != nil {
			fmt.Printf("Error writing file for package %s: %s\n", filename, err.Error())
			os.Exit(1)
		}
	}

}
