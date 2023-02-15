package ubi8nodeenginebuildpackextension

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/draft"
)

//go:embed templates/build.Dockerfile
var buildDockerfileTemplate string

//go:embed templates/run.Dockerfile
var runDockerfileTemplate string

type BuildDockerfileProps struct {
	NODEJS_VERSION, CNB_USER_ID, CNB_GROUP_ID int
	CNB_STACK_ID, PACKAGES                    string
}

type RunDockerfileProps struct {
	Registry       string
	NODEJS_VERSION int
}

// from nodejs-engine buildpack, keep in sync
var priorities = []interface{}{
	"BP_NODE_VERSION",
	"package.json",
	".nvmrc",
	".node-version",
}

var defaultNodejsVersion = 18

func ReadAndDecodeBuildpackPlan(buildpackPlanPath string) (buildpackPlan packit.BuildpackPlan, Error error) {
	var plan packit.BuildpackPlan
	var _, err = toml.DecodeFile(buildpackPlanPath, &plan)

	if err != nil {
		return plan, err
	}
	return plan, nil
}

func ResolveNodeVersionByPriorities(plan packit.BuildpackPlan, priorities []interface{}) (version string, Error error) {

	entryResolver := draft.NewPlanner()

	entry, _ := entryResolver.Resolve("node", plan.Entries, priorities)
	if entry.Name == "" {
		return "", errors.New("No Node.js version found.")
	}

	if version, ok := entry.Metadata["version"].(string); ok {
		return version, nil
	} else {
		return "", nil
	}
}

func ResolveNodeVersionConstrains(version string, defaultNodejsVersion int) (nodejsVersion int, Error error) {
	if version == "" {
		return defaultNodejsVersion, nil
	}

	constraint, err := semver.NewConstraint(version)
	if err != nil {

		return 0, errors.New("Could not parse Node.js version")
	}

	// we should make this check as close as possible to what
	// is in the dependency resolve which is more forgiving
	// than this. The versions should also be set by what
	// the actual version numbers in the build image
	version18, _ := semver.NewVersion("18")

	if constraint.Check(version18) {
		return 18, nil
	}

	version16, _ := semver.NewVersion("16")

	if constraint.Check(version16) {
		return 16, nil
	}

	return 0, errors.New("Unsupported Node.js version")
}

func Generate(config OptionConfig) {

	var planPath = os.Getenv("CNB_BP_PLAN_PATH")

	plan, err := ReadAndDecodeBuildpackPlan(planPath)

	if err != nil {
		fmt.Println(err)
		config.ExitHandler.Error(Fail)
		return
	}

	version, err := ResolveNodeVersionByPriorities(plan, priorities)

	if err != nil {
		config.ExitHandler.Error(Fail)
		return
	}

	// Extract the version of Node.js to install, default is 18
	// This logic will vary based on what is supported by the ubi image
	NODEJS_VERSION, err := ResolveNodeVersionConstrains(version, defaultNodejsVersion)

	if err != nil {
		fmt.Println(err)
		config.ExitHandler.Error(Fail)
		return
	}

	fmt.Println("VERSION:", NODEJS_VERSION)

	// Below variables has to be fetch from the env
	CNB_PLATFORM_API := os.Getenv("CNB_PLATFORM_API")
	CNB_STACK_ID := os.Getenv("CNB_STACK_ID")

	// INPUT ARGUMENTS
	outputDir := config.Args[1]

	//  Patched by build.sh with correct values
	CNB_USER_ID := 1000
	CNB_GROUP_ID := 1000

	fmt.Println("****************************")

	fmt.Println("extension build env vars!!")
	fmt.Println("CNB_PLATFORM_API:", CNB_PLATFORM_API)
	fmt.Println("CNB_STACK_ID: ", CNB_STACK_ID)
	fmt.Println("CNB_USER_ID: ", CNB_USER_ID)
	fmt.Println("CNB_GROUP_ID: ", CNB_GROUP_ID)

	fmt.Println("****************************")

	fmt.Println("extension plan...")

	err = readFileAndPrintToStdout(planPath)
	if err != nil {
		config.ExitHandler.Error(Fail)
		return
	}

	fmt.Println("****************************")

	//  TODO .. read engines from $3 to select
	//          appropriate rpm
	//          for PoC purposes a single nodejs version will do.
	//          currently hard coded to 16
	//  Search for any tools on parsing the .toml files.
	//  Check how to use a binary file generated from go

	/* Creating build.Dockerfile*/

	buildDockerfileContent := bytes.Buffer{}

	builDockerfileProps := BuildDockerfileProps{
		NODEJS_VERSION: NODEJS_VERSION,
		CNB_USER_ID:    CNB_USER_ID,
		CNB_GROUP_ID:   CNB_GROUP_ID,
		CNB_STACK_ID:   CNB_STACK_ID,
		PACKAGES:       "make gcc gcc-c++ libatomic_ops git openssl-devel nodejs npm nodejs-nodemon nss_wrapper which",
	}

	err = generateDockerfileFromTemplate(&buildDockerfileContent, builDockerfileProps, buildDockerfileTemplate)

	if err != nil {
		config.ExitHandler.Error(Fail)
		return
	}

	writeContentToFile(buildDockerfileContent.String(), outputDir+"/build.Dockerfile")

	if err != nil {
		config.ExitHandler.Error(Fail)
		return
	}

	/* Creating run.Dockerfile */

	runDockerfileContent := bytes.Buffer{}

	runDockerfileProps := RunDockerfileProps{
		Registry:       "172.17.0.1:5000",
		NODEJS_VERSION: NODEJS_VERSION,
	}

	err = generateDockerfileFromTemplate(&runDockerfileContent, runDockerfileProps, runDockerfileTemplate)

	if err != nil {
		config.ExitHandler.Error(Fail)
		return
	}

	err = writeContentToFile(runDockerfileContent.String(), outputDir+"/run.Dockerfile")
	if err != nil {
		config.ExitHandler.Error(Fail)
		return
	}

	fmt.Println("Output of build and run Dockerfiles complete")

	cmd := exec.Command("ls", "-al", outputDir)
	stdout, err := cmd.Output()

	if err != nil {
		config.ExitHandler.Error(Fail)
		return
	}
	fmt.Print(string(stdout))

}

func generateDockerfileFromTemplate(w io.Writer, dockerfileProps interface{}, dockerfileTemplate string) error {

	templ, err := template.New("dockerfile").Parse(dockerfileTemplate)

	if err != nil {
		return err
	}

	if err := templ.Execute(w, dockerfileProps); err != nil {
		return err
	}

	return nil
}

func writeContentToFile(fileContent string, filepath string) (Error error) {

	f, err := os.Create(filepath)

	if err != nil {
		return err
	}

	n, err := f.WriteString(fileContent)
	if err != nil {
		return err
	}
	fmt.Println(n)
	f.Sync()

	return nil
}

func readFileAndPrintToStdout(filepath string) error {

	filecontent, err := os.ReadFile(filepath)

	if err != nil {
		return err
	}

	fmt.Println(string(filecontent))

	return nil
}
