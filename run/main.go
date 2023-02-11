package main

import (
	"github.com/paketo-buildpacks/packit/v2"

	ubi8nodeenginebuildpackextension "github.com/nodeshift/ubi8-node-engine-buildack-extension"
)

func main() {
	packit.Run(
		ubi8nodeenginebuildpackextension.Detect(),
		ubi8nodeenginebuildpackextension.Generate(),
	)
}
