package ubi8nodeenginebuildpackextension

import (
	"fmt"
	"os"
	"path/filepath"

	nodestart "github.com/paketo-buildpacks/node-start"
	npmstart "github.com/paketo-buildpacks/npm-start"
)

// functionality from npm-start buildpack, also some overlap with npm-install
func packageJSONExists(workingDir string, projectPathParser npmstart.PathParser) (path string, err error) {

	projectPath, err := projectPathParser.Get(workingDir)
	if err != nil {
		return "", err
	}

	path = filepath.Join(projectPath, "package.json")
	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return path, nil
}

// functionality from node-start
func nodeApplicationExists(workingDir string, applicationFinder nodestart.ApplicationFinder) (path string, err error) {
	return applicationFinder.Find(workingDir, os.Getenv("BP_LAUNCHPOINT"), os.Getenv("BP_NODE_PROJECT_PATH"))
}

// ExitHandler serves as the interface for types that can handle an error
// during the Detect or Build functions. ExitHandlers are responsible for
// translating error values into exit codes according the specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#detection and
// https://github.com/buildpacks/spec/blob/main/buildpack.md#build.
//
//go:generate faux --interface ExitHandlerInterface --output fakes/exit_handler.go
type ExitHandlerInterface interface {
	Error(err error)
}

func Detect(config OptionConfig) {
	planPath := config.Args[2]

	// likely move these to main.go ?
	workingDir, err := os.Getwd()
	if err != nil {
		config.ExitHandler.Error(Fail)
		return
	}
	projectPathParser := npmstart.NewProjectPathParser()
	nodeApplicationFinder := nodestart.NewNodeApplicationFinder()

	fmt.Printf("Extension detect, with plan path %s", planPath)

	packageJSON, err := packageJSONExists(workingDir, projectPathParser)
	if err != nil {
		config.ExitHandler.Error(Fail)
		return
	}

	if packageJSON == "" {
		// no package.json so look for know Node.js application files
		path, err := nodeApplicationExists(workingDir, nodeApplicationFinder)
		if err != nil {
			config.ExitHandler.Error(Fail)
			return
		}
		// if no application was found then we don't need to provide node
		if path == "" {
			config.ExitHandler.Error(Fail)
			return
		}
	}

	// if we get here we either found a packge.json or Node.js application file
	fmt.Println("Node.js extension adding to build plan")
	content := `[[provides]]
    name = "node"
   
    [[or]]
    [[or.provides]]
    name = "node"
    [[or.provides]]
    name = "npm"`

	_, err = appendContentTofile(planPath, content)
	if err != nil {
		config.ExitHandler.Error(Fail)
		return
	}
	config.ExitHandler.Error(nil)
}

func appendContentTofile(filename string, content string) (string, error) {

	f, err := os.OpenFile(filename,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return "", err
	}

	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return "", err
	}

	return "ok", nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
