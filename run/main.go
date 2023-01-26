package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	ubi8nodeenginebuildpackextension "github.com/nodeshift/ubi8-node-engine-buildack-extension"
	"github.com/nodeshift/ubi8-node-engine-buildack-extension/internal"
)

func main() {

	run(ubi8nodeenginebuildpackextension.Detect, ubi8nodeenginebuildpackextension.Generate)
}

type fn func()

// Run combines the invocation of both build and detect into a single entry
// point. Calling Run from an executable with a name matching "build" or
// "detect" will result in the matching DetectFunc or BuildFunc being called.
func run(Detect fn, Generate fn, options ...Option) {
	config := OptionConfig{
		exitHandler: internal.NewExitHandler(),
		args:        os.Args,
	}

	for _, option := range options {
		config = option(config)
	}

	fmt.Println(config)
	phase := filepath.Base(config.args[0])

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
		Detect()

	case "generate":
		Generate()

	default:
		config.exitHandler.Error(fmt.Errorf("failed to run buildpack: unknown lifecycle phase %q", phase))
	}

}

// //////////////////////Options//////////////////////////
// OptionConfig is the set of configurable options for the Build and Detect
// functions.
type OptionConfig struct {
	exitHandler ExitHandler
	args        []string
	tomlWriter  TOMLWriter
	envWriter   EnvironmentWriter
	fileWriter  FileWriter
}

// Option declares a function signature that can be used to define optional
// modifications to the behavior of the Detect and Build functions.
type Option func(config OptionConfig) OptionConfig

//go:generate faux --interface ExitHandler --output fakes/exit_handler.go

// ExitHandler serves as the interface for types that can handle an error
// during the Detect or Build functions. ExitHandlers are responsible for
// translating error values into exit codes according the specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#detection and
// https://github.com/buildpacks/spec/blob/main/buildpack.md#build.
type ExitHandler interface {
	Error(error)
}

// TOMLWriter serves as the interface for types that can handle the writing of
// TOML files. TOMLWriters take a path to a file location on disk and a
// datastructure to marshal.
type TOMLWriter interface {
	Write(path string, value interface{}) error
}

// EnvironmentWriter serves as the interface for types that can write an
// Environment to a directory on disk according to the specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#provided-by-the-buildpacks.
type EnvironmentWriter interface {
	Write(dir string, env map[string]string) error
}

type FileWriter interface {
	Write(path string, reader io.Reader) error
}

// WithExitHandler is an Option that overrides the ExitHandler for a given
// invocation of Build or Detect.
func WithExitHandler(exitHandler ExitHandler) Option {
	return func(config OptionConfig) OptionConfig {
		config.exitHandler = exitHandler
		return config
	}
}

// WithArgs is an Option that overrides the value of os.Args for a given
// invocation of Build or Detect.
func WithArgs(args []string) Option {
	return func(config OptionConfig) OptionConfig {
		config.args = args
		return config
	}
}
