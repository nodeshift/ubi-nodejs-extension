package ubi8nodeenginebuildpackextension_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	ubi8nodeenginebuildpackextension "github.com/nodeshift/ubi8-node-engine-buildack-extension"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	"github.com/BurntSushi/toml"

	fakes "github.com/nodeshift/ubi8-node-engine-buildack-extension/fakes"
)

//go:embed testdata/testdata.build.Dockerfile
var outputBuildDockerfile string

//go:embed testdata/testdata.run.Dockerfile
var outputRunDockerfile string

func testGenerate(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect      = NewWithT(t).Expect
		exitHandler fakes.ExitHandlerInterface
		config      ubi8nodeenginebuildpackextension.OptionConfig
		workingDir  string
		// platformDir   string
		outputDir     string
		planPath      string
		testBuildPlan packit.BuildPlan
		buf           = new(bytes.Buffer)
	)

	context("Generate called with NO node in buildplan", func() {
		it.Before(func() {
			workingDir = t.TempDir()
			// platformDir = t.TempDir()
			outputDir = t.TempDir()

			testBuildPlan = packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					// {Name: "node"},
				},
			}

			err := toml.NewEncoder(buf).Encode(testBuildPlan)
			fmt.Print(err)

			Expect(os.WriteFile(filepath.Join(workingDir, "plan"), buf.Bytes(), 0600)).To(Succeed())
			// planPath = filepath.Join(workingDir, "plan")

			exitHandler = fakes.ExitHandlerInterface{}
			config = ubi8nodeenginebuildpackextension.OptionConfig{
				ExitHandler: &exitHandler,
				Args:        []string{"exe", outputDir},
			}
			os.Chdir(workingDir)
		})

		it("Node no longer requested in buildplan", func() {
			ubi8nodeenginebuildpackextension.Generate(config)
			Expect(exitHandler.ErrorCall.Receives.Err.Error()).To(Equal("failed"))
			// writeContentToFile(buildDockerfileContent, outputDir+"/build.Dockerfile")
		})
	}, spec.Sequential())

	context("Generate called with node in the buildplan", func() {
		it.Before(func() {

			workingDir = t.TempDir()
			// platformDir = t.TempDir()
			outputDir = t.TempDir()

			testBuildPlan := packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "node",
					},
				},
			}

			err := toml.NewEncoder(buf).Encode(testBuildPlan)
			fmt.Print(err)
			planPath = filepath.Join(workingDir, "plan")
			t.Setenv("CNB_BP_PLAN_PATH", planPath)

			Expect(os.WriteFile(planPath, buf.Bytes(), 0600)).To(Succeed())

			exitHandler = fakes.ExitHandlerInterface{}
			config = ubi8nodeenginebuildpackextension.OptionConfig{
				ExitHandler: &exitHandler,
				Args:        []string{"exe", outputDir},
			}
			os.Chdir(workingDir)
		})

		it("Node specific version of node requested", func() {
			ubi8nodeenginebuildpackextension.Generate(config)
			Expect(exitHandler.ErrorCall.Receives.Err).To(BeNil())
		})

		it("should generate build and run docker files.", func() {
			ubi8nodeenginebuildpackextension.Generate(config)

			buildDockerfileFilepath := outputDir + "/build.Dockerfile"
			buildDockerfile, _ := os.ReadFile(buildDockerfileFilepath)
			Expect(outputBuildDockerfile).To(Equal(string(buildDockerfile)))

			runDockerfileFilepath := outputDir + "/run.Dockerfile"
			runDockerfile, _ := os.ReadFile(runDockerfileFilepath)
			Expect(outputRunDockerfile).To(Equal(string(runDockerfile)))

		})
	}, spec.Sequential())

	context("Resolution of Node.js version based on priorities", func() {

		it("should properly resolve Node.js versions.", func() {

			nodeVersionTests := []struct {
				priorities     []interface{}
				testBuildPlan  packit.BuildpackPlan
				hasNodeVersion string
			}{
				{priorities: []interface{}{
					"BP_NODE_VERSION",
					"package.json",
					".nvmrc",
					".node-version",
				}, testBuildPlan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "node",
							Metadata: map[string]interface{}{
								"version":        "~14",
								"version-source": "BP_NODE_VERSION",
							},
						},
					},
				}, hasNodeVersion: "~14"},
				{priorities: []interface{}{
					"BP_NODE_VERSION",
					"package.json",
					".nvmrc",
					".node-version",
				}, testBuildPlan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "node",
							Metadata: map[string]interface{}{
								"version":        "~16",
								"version-source": "BP_NODE_VERSION",
							},
						},
					},
				}, hasNodeVersion: "~16"},
				{priorities: []interface{}{
					"BP_NODE_VERSION",
					"package.json",
					".nvmrc",
					".node-version",
				}, testBuildPlan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "node",
							Metadata: map[string]interface{}{
								"version":        "~18",
								"version-source": "BP_NODE_VERSION",
							},
						},
					},
				}, hasNodeVersion: "~18"},
			}

			for _, tt := range nodeVersionTests {
				version, _ := ubi8nodeenginebuildpackextension.ResolveNodeVersionByPriorities(tt.testBuildPlan, tt.priorities)
				Expect(tt.hasNodeVersion).To(Equal(version))
			}

		})
	}, spec.Sequential())

	context("Reading and decoding buildpackplan", func() {

		it.Before(func() {
			workingDir = t.TempDir()
			planPath = filepath.Join(workingDir, "plan")
		})

		it("should properly parse and return buildPlan", func() {

			fileDecodingTests := []struct {
				planPath          string
				testBuildpackPlan packit.BuildpackPlan
			}{

				{planPath: planPath, testBuildpackPlan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "node",
						},
					},
				}},
			}

			for _, tt := range fileDecodingTests {

				err := toml.NewEncoder(buf).Encode(tt.testBuildpackPlan)
				fmt.Print(err)

				Expect(os.WriteFile(planPath, buf.Bytes(), 0600)).To(Succeed())

				decodedBuildPlan, _ := ubi8nodeenginebuildpackextension.ReadAndDecodeBuildpackPlan(tt.planPath)
				Expect(decodedBuildPlan).To(Equal(tt.testBuildpackPlan))
			}

		})
	}, spec.Sequential())

}
