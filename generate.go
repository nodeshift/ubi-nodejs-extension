package ubi8nodeenginebuildpackextension

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/draft"
)

func Generate() {

	// Extract the version of Node.js to install, default is 18
	// This logic will vary based on what is supported by the ubi image
	NODEJS_VERSION := 18
	entryResolver := draft.NewPlanner()

	var plan packit.BuildpackPlan
	fmt.Println(plan)
	var _, err = toml.DecodeFile(os.Getenv("CNB_BP_PLAN_PATH"), &plan)

	// from nodejs-engine buildpack, keep in sync
	priorities := []interface{}{
		"BP_NODE_VERSION",
		"package.json",
		".nvmrc",
		".node-version",
	}

	entry, _ := entryResolver.Resolve("node", plan.Entries, priorities)
	version := entry.Metadata["version"]

	if version != nil && version != "" {
		constraint, err := semver.NewConstraint(version.(string))
		if err != nil {
			// Handle constraint not being parseable.
			fmt.Println("Could not parse Node.js version")
			os.Exit(100)
		}

		// we should make this check as close as possible to what
		// is in the dependency resolve which is more forgiving
		// than this. The versions should also be set by what
		// the actual version numbers in the build image
		version18, _ := semver.NewVersion("18")
		version16, _ := semver.NewVersion("16")
		if constraint.Check(version18) {
			NODEJS_VERSION = 18
		} else if constraint.Check(version16) {
			NODEJS_VERSION = 16
		} else {
			fmt.Println("Unsupported Node.js version")
			os.Exit(100)
		}
	}
	fmt.Println("VERSION:", NODEJS_VERSION)

	// Below variables has to be fetch from the env
	CNB_PLATFORM_API := os.Getenv("CNB_PLATFORM_API")
	CNB_STACK_ID := os.Getenv("CNB_STACK_ID")

	// INPUT ARGUMENTS
	fmt.Println("--->", os.Args)
	platformDir := os.Args[2]
	envDir := platformDir + "/env"
	outputDir := os.Args[1]
	planPath := os.Args[3]

	//  Patched by build.sh with correct values
	CNB_USER_ID := 1000
	CNB_GROUP_ID := 1000

	fmt.Println("GO****************************")
	fmt.Println("ouput_dir", outputDir)
	fmt.Println("plan_path", planPath)
	fmt.Println("env_dir", envDir)
	fmt.Println("extension build env vars!!")
	fmt.Println("CNB_PLATFORM_API:", CNB_PLATFORM_API)
	fmt.Println("CNB_STACK_ID: ", CNB_STACK_ID)
	fmt.Println("CNB_USER_ID: ", CNB_USER_ID)
	fmt.Println("CNB_GROUP_ID: ", CNB_GROUP_ID)
	fmt.Println("****************************")
	fmt.Println("extension plan...")
	readFileAndPrintToStdout(planPath)
	fmt.Println("****************************")

	// return
	//  TODO .. read engines from $3 to select
	//          appropriate rpm
	//          for PoC purposes a single nodejs version will do.
	//          currently hard coded to 16
	//  Search for any tools on parsing the .toml files.
	//  Check how to use a binary file generated from go

	PACKAGES := "make gcc gcc-c++ libatomic_ops git openssl-devel nodejs npm nodejs-nodemon nss_wrapper which"

	buildDockerfileContent := fmt.Sprintf(`ARG base_image 
FROM ${base_image}

USER root
ARG build_id=0
RUN echo ${build_id}
	
RUN microdnf -y module enable nodejs:%d`, NODEJS_VERSION)
	buildDockerfileContent += fmt.Sprintf("\n")

	buildDockerfileContent += fmt.Sprintf(`RUN microdnf --setopt=install_weak_deps=0 --setopt=tsflags=nodocs install -y %s && microdnf clean all`, PACKAGES)
	buildDockerfileContent += fmt.Sprintf("\n\n")

	buildDockerfileContent += fmt.Sprintf(`RUN echo uid:gid "%d:%d"`, CNB_USER_ID, CNB_GROUP_ID)
	buildDockerfileContent += fmt.Sprintf("\n")

	buildDockerfileContent += fmt.Sprintf(`USER %d:%d`, CNB_USER_ID, CNB_GROUP_ID)
	buildDockerfileContent += fmt.Sprintf("\n\n")

	buildDockerfileContent += fmt.Sprintf(`RUN echo "CNB_STACK_ID: %s"`, CNB_STACK_ID)
	buildDockerfileContent += fmt.Sprintf("\n")

	writeContentToFile(buildDockerfileContent, outputDir+"/build.Dockerfile")

	// default is 18
	runDockerfileContent := "FROM 172.17.0.1:5000/ubi8-paketo-run-nodejs-18"
	if NODEJS_VERSION == 16 {
		runDockerfileContent = "FROM 172.17.0.1:5000/ubi8-paketo-run-nodejs-16"
	}

	writeContentToFile(runDockerfileContent, outputDir+"/run.Dockerfile")

	// fmt.Println("===>", runDockerfileContent, "<===")

	fmt.Println("Output of build and run Dockerfiles complete")
	cmd := exec.Command("ls", "-al", outputDir)
	stdout, err := cmd.Output()

	//TODO return something that will exit the whole process
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Print(string(stdout))
}

func writeContentToFile(fileContent string, filepath string) {

	f, err := os.Create(filepath)

	//TODO return something that will exit the whole process
	if err != nil {
		log.Fatal(err)
		return
	}

	n, err := f.WriteString(fileContent + "\n")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(n)
	f.Sync()
}

func readFileAndPrintToStdout(filepath string) {

	f, err := os.Open(filepath)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	reader := bufio.NewReader(f)
	buf := make([]byte, 16)

	for {
		n, err := reader.Read(buf)

		if err != nil {

			if err != io.EOF {

				log.Fatal(err)
			}

			break
		}

		fmt.Print(string(buf[0:n]))
	}

	fmt.Println()
}
