package main

import (
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/postal"

	ubi8nodeenginebuildpackextension "github.com/nodeshift/ubi8-node-engine-buildack-extension"
)

func main() {
	dependencyManager := postal.NewService(cargo.NewTransport())

	packit.RunExtension(
		ubi8nodeenginebuildpackextension.Detect(),
		ubi8nodeenginebuildpackextension.Generate(dependencyManager),
	)
}
