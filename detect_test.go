package ubi8nodeenginebuildpackextension_test

import (
	"testing"

	ubi8nodeenginebuildpackextension "github.com/nodeshift/ubi8-node-engine-buildack-extension"
	fakes "github.com/nodeshift/ubi8-node-engine-buildack-extension/fakes"

	//	"github.com/paketo-buildpacks/packit/v2"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect      = NewWithT(t).Expect
		exitHandler fakes.ExitHandlerInterface
		config      ubi8nodeenginebuildpackextension.OptionConfig
		workingDir  string
		planPath    string
	)

	context("when no application is detected", func() {
		it.Before(func() {
			workingDir = t.TempDir()
			Expect(os.MkdirAll(filepath.Join(workingDir, "src"), os.ModePerm)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(workingDir, "plan"), nil, 0600)).To(Succeed())
			planPath = filepath.Join(workingDir, "plan")

			exitHandler = fakes.ExitHandlerInterface{}
			config = ubi8nodeenginebuildpackextension.OptionConfig{
				ExitHandler: &exitHandler,
				Args:        []string{"exe", "arg1", planPath},
			}
			os.Chdir(workingDir)
		})

		it("indicates it does not participate", func() {
			ubi8nodeenginebuildpackextension.Detect(config)
			Expect(exitHandler.ErrorCall.Receives.Err.Error()).To(Equal("failed"))
		})
	}, spec.Sequential())

	context("when an application is detected in the working dir", func() {
		it.Before(func() {
			workingDir = t.TempDir()
			t.Setenv("BP_NODE_PROJECT_PATH", "./src")
			t.Setenv("BP_LAUNCHPOINT", "./src/server.js")
			Expect(os.MkdirAll(filepath.Join(workingDir, "src"), os.ModePerm)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(workingDir, "src", "server.js"), nil, 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(workingDir, "plan"), nil, 0600)).To(Succeed())
			planPath = filepath.Join(workingDir, "plan")

			exitHandler = fakes.ExitHandlerInterface{}
			config = ubi8nodeenginebuildpackextension.OptionConfig{
				ExitHandler: &exitHandler,
				Args:        []string{"exe", "arg1", planPath},
			}
			os.Chdir(workingDir)
		})

		it("detects", func() {
			ubi8nodeenginebuildpackextension.Detect(config)
			var plan packit.BuildPlan
			var _, _ = toml.DecodeFile(planPath, &plan)
			Expect(exitHandler.ErrorCall.Receives.Err).To(BeNil())
			Expect(plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "node"},
				},
				Or: []packit.BuildPlan{
					{
						Provides: []packit.BuildPlanProvision{
							{Name: "node"},
							{Name: "npm"},
						},
					},
				},
			}))
		})
	}, spec.Sequential())
}
