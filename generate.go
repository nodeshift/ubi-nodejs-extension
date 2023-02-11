package ubi8nodeenginebuildpackextension

import (
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/draft"
)

func Generate() packit.GenerateFunc {
	return func(context packit.GenerateContext) (packit.GenerateResult, error) {

		// likely move this out to main
		entryResolver := draft.NewPlanner()

		// Default version of Node.js to install
		NODEJS_VERSION := 18

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

		version := entry.Metadata["version"]

		if version != nil && version != "" {
			constraint, err := semver.NewConstraint(version.(string))
			if err != nil {
				// Handle constraint not being parseable.
				return packit.GenerateResult{}, packit.Fail.WithMessage("Could not parse Node.js version")
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
				return packit.GenerateResult{}, packit.Fail.WithMessage("Unsupported Node.js version")
			}
		}
		fmt.Println("VERSION:", NODEJS_VERSION)

		// Below variables has to be fetch from the env
		// CNB_PLATFORM_API := os.Getenv("CNB_PLATFORM_API")
		CNB_STACK_ID := os.Getenv("CNB_STACK_ID")

		//  Should be externalized
		CNB_USER_ID := 1000
		CNB_GROUP_ID := 1000

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

		// default is 18
		runDockerfileContent := "FROM 172.17.0.1:5000/ubi8-paketo-run-nodejs-18"
		if NODEJS_VERSION == 16 {
			runDockerfileContent = "FROM 172.17.0.1:5000/ubi8-paketo-run-nodejs-16"
		}

		return packit.GenerateResult{
			packit.ExtendConfig{Build: packit.ExtendImageConfig{[]packit.ExtendImageConfigArg{}}},
			strings.NewReader(buildDockerfileContent),
			strings.NewReader(runDockerfileContent),
		}, nil
	}
}
