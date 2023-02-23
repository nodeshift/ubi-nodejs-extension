package ubi8nodeenginebuildpackextension

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"
	"path/filepath"
	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/draft"
	postal "github.com/paketo-buildpacks/packit/v2/postal"
)

//go:embed templates/build.Dockerfile
var buildDockerfileTemplate string

type BuildDockerfileProps struct {
	NODEJS_VERSION            uint64
	CNB_USER_ID, CNB_GROUP_ID int
	CNB_STACK_ID, PACKAGES    string
}


//go:embed templates/run.Dockerfile
var runDockerfileTemplate string

type RunDockerfileProps struct {
	Source string
}

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Deliver(dependency postal.Dependency, cnbPath, layerPath, platformPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}


func Generate(dependencyManager DependencyManager) packit.GenerateFunc {
	return func(context packit.GenerateContext) (packit.GenerateResult, error) {

		// likely move this out to main
		entryResolver := draft.NewPlanner()

		// from nodejs-engine buildpack, keep in sync
		priorities := []interface{}{
			"BP_NODE_VERSION",
			"package.json",
			".nvmrc",
			".node-version",
		}

		entry, _ := entryResolver.Resolve("node", context.Plan.Entries, priorities)
		if entry.Name == "" {
			return packit.GenerateResult{}, packit.Fail.WithMessage("Node.js no longer requested by build plan")
		}

		version, _ := entry.Metadata["version"].(string)
		extensionFilePath := filepath.Join(context.CNBPath, "extension.toml")
		dependency, err := dependencyManager.Resolve(extensionFilePath, entry.Name, version, context.Stack)
		if err != nil {
			return packit.GenerateResult{}, err
		}

		sVersion, err := semver.NewVersion(dependency.Version)


		NODEJS_VERSION := sVersion.Major()
		fmt.Println("NODEJS MAJOR VERSION:", NODEJS_VERSION)

		// Below variables has to be fetch from the env
		// CNB_PLATFORM_API := os.Getenv("CNB_PLATFORM_API")
		CNB_STACK_ID := os.Getenv("CNB_STACK_ID")

		//  Should be externalized
		CNB_USER_ID := 1000
		CNB_GROUP_ID := 1000

		fmt.Println("****************************")

		fmt.Println("extension build env vars!!")
		// fmt.Println("CNB_PLATFORM_API:", CNB_PLATFORM_API)
		fmt.Println("CNB_STACK_ID: ", CNB_STACK_ID)
		fmt.Println("CNB_USER_ID: ", CNB_USER_ID)
		fmt.Println("CNB_GROUP_ID: ", CNB_GROUP_ID)

		fmt.Println("****************************")


		fmt.Println("extension plan...")

		/* Creating build.Dockerfile*/

		builDockerfileProps := BuildDockerfileProps{
			NODEJS_VERSION: NODEJS_VERSION,
			CNB_USER_ID:    CNB_USER_ID,
			CNB_GROUP_ID:   CNB_GROUP_ID,
			CNB_STACK_ID:   CNB_STACK_ID,
			PACKAGES:       "make gcc gcc-c++ libatomic_ops git openssl-devel nodejs npm nodejs-nodemon nss_wrapper which",
		}

		buildDockerfileContent, err := FillPropsToTemplate(builDockerfileProps, buildDockerfileTemplate)

		if err != nil {
			return packit.GenerateResult{}, err
		}

		/* Creating run.Dockerfile*/

		RunDockerfileProps := RunDockerfileProps{
			Source: dependency.Source,
		}

		runDockerfileContent, err := FillPropsToTemplate(RunDockerfileProps, runDockerfileTemplate)

		if err != nil {
			return packit.GenerateResult{}, err
		}

		return packit.GenerateResult{
			ExtendConfig:    packit.ExtendConfig{Build: packit.ExtendImageConfig{[]packit.ExtendImageConfigArg{}}},
			BuildDockerfile: strings.NewReader(buildDockerfileContent),
			RunDockerfile:   strings.NewReader(runDockerfileContent),
		}, nil
	}
}

func FillPropsToTemplate(properties interface{}, templateString string) (result string, Error error) {

	templ, err := template.New("template").Parse(templateString)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = templ.Execute(&buf, properties)
	if err != nil {
		panic(err)
	}

	return buf.String(), nil

}
