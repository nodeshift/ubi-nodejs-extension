package ubi8nodeenginebuildpackextension

// package ubi8nodeenginebuildpackextension

import (
	"fmt"
	"os"
)

func Detect() {
	planPath := os.Args[2]

	fmt.Printf("Extension detect, with plan path %s", planPath)

	// DOC: Here, we need to know if we should supply a node/npm etc..
	// we have to do the project determination here, because we run before
	// any buildpack detect will.. and we need to say what we will 'provide'
	// For now, we'll just look for hallmarks of a node.js project, and if found
	// claim we will 'provide' a node/npm

	if fileExists("package.json") {

		fmt.Println("Node.js extension adding to build plan")
		content := `[[provides]]
    name = "node"
   
    [[or]]
    [[or.provides]]
    name = "node"
    [[or.provides]]
    name = "npm"`

		_, err := appendContentTofile(planPath, content)
		if err != nil {
			os.Exit(100)
			return
		}
		os.Exit(0)

	} else {
		fmt.Println("sample file does not exist")
		os.Exit(100)
	}
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
