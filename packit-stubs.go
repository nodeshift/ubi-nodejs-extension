package ubi8nodeenginebuildpackextension

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// //////////////////////Options//////////////////////////
// OptionConfig is the set of configurable options for the Build and Detect
// functions.
type OptionConfig struct {
	ExitHandler ExitHandlerInterface
	Args        []string
	tomlWriter  TOMLWriter
	envWriter   EnvironmentWriter
	fileWriter  FileWriter
}

// Option declares a function signature that can be used to define optional
// modifications to the behavior of the Detect and Build functions.
type Option func(config OptionConfig) OptionConfig

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
		config.ExitHandler = exitHandler
		return config
	}
}

// WithArgs is an Option that overrides the value of os.Args for a given
// invocation of Build or Detect.
func WithArgs(args []string) Option {
	return func(config OptionConfig) OptionConfig {
		config.Args = args
		return config
	}
}

var Fail = failError{error: errors.New("failed")}

type failError struct {
	error
}

func (f failError) WithMessage(format string, v ...interface{}) failError {
	return failError{error: fmt.Errorf(format, v...)}
}

type Option2 func(handler ExitHandler) ExitHandler

func WithExitHandlerStderr(stderr io.Writer) Option2 {
	return func(handler ExitHandler) ExitHandler {
		handler.stderr = stderr
		return handler
	}
}

func WithExitHandlerStdout(stdout io.Writer) Option2 {
	return func(handler ExitHandler) ExitHandler {
		handler.stdout = stdout
		return handler
	}
}

func WithExitHandlerExitFunc(e func(int)) Option2 {
	return func(handler ExitHandler) ExitHandler {
		handler.exitFunc = e
		return handler
	}
}

type ExitHandler struct {
	stdout   io.Writer
	stderr   io.Writer
	exitFunc func(int)
	ExitHandlerInterface
}

func NewExitHandler(options ...Option2) ExitHandler {
	handler := ExitHandler{
		stdout:   os.Stdout,
		stderr:   os.Stderr,
		exitFunc: os.Exit,
	}

	for _, option := range options {
		handler = option(handler)
	}

	return handler
}

func (h ExitHandler) Error(err error) {
	fmt.Fprintln(h.stderr, err)

	var code int
	switch err.(type) {
	case failError:
		code = 100
	case nil:
		code = 0
	default:
		code = 1
	}

	h.exitFunc(code)
}
