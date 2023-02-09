package ubi8nodeenginebuildpackextension_test

import (
	"bytes"
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
			// writeContentToFile(buildDockerfileContent, outputDir+"/build.Dockerfile")
		})
	}, spec.Sequential())

}
