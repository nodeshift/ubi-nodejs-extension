package main

import (
	"fmt"
	"os"
	"path/filepath"

	ubi8nodeenginebuildpackextension "github.com/nodeshift/ubi8-node-engine-buildack-extension"
)

func main() {

	run(ubi8nodeenginebuildpackextension.Detect, ubi8nodeenginebuildpackextension.Generate)
}

type fn func()

type DetectFunction func(options ubi8nodeenginebuildpackextension.OptionConfig)
type GenerateFunction func(options ubi8nodeenginebuildpackextension.OptionConfig)

// Run combines the invocation of both build and detect into a single entry
// point. Calling Run from an executable with a name matching "build" or
// "detect" will result in the matching DetectFunc or BuildFunc being called.
func run(Detect DetectFunction, Generate GenerateFunction, options ...ubi8nodeenginebuildpackextension.Option) {
	config := ubi8nodeenginebuildpackextension.OptionConfig{
		ExitHandler: ubi8nodeenginebuildpackextension.NewExitHandler(),
		Args:        os.Args,
	}
	for _, option := range options {
		config = option(config)
	}

	fmt.Println(config)
	phase := filepath.Base(config.Args[0])

	fmt.Println("**Start of Run function*******")
	fmt.Println("**Phase*******")
	fmt.Println(phase)
	fmt.Println("**Options*******")
	fmt.Println(options)
	fmt.Println("**Config*******")
	fmt.Println(config)
	fmt.Println("**end of Run function*******")

	switch phase {
	case "detect":
		Detect(config)

	case "generate":
		Generate(config)

	default:
		config.ExitHandler.Error(fmt.Errorf("failed to run buildpack: unknown lifecycle phase %q", phase))
	}

}
