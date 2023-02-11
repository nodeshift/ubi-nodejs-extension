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
)

func testGenerate(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect         = NewWithT(t).Expect
		workingDir     string
		planPath       string
		testBuildPlan  packit.BuildpackPlan
		buf            = new(bytes.Buffer)
		generateResult packit.GenerateResult
		err            error
	)

	context("Generate called with NO node in buildplan", func() {
		it.Before(func() {
			workingDir = t.TempDir()

			err := toml.NewEncoder(buf).Encode(testBuildPlan)
			fmt.Print(err)

			Expect(os.WriteFile(filepath.Join(workingDir, "plan"), buf.Bytes(), 0600)).To(Succeed())
			// planPath = filepath.Join(workingDir, "plan")

			os.Chdir(workingDir)
		})

		it("Node no longer requested in buildplan", func() {
			generateResult, err = ubi8nodeenginebuildpackextension.Generate()(packit.GenerateContext{
				WorkingDir: workingDir,
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
			})
			Expect(err).To(HaveOccurred())
			Expect(generateResult.BuildDockerfile).To(BeNil())
			// writeContentToFile(buildDockerfileContent, outputDir+"/build.Dockerfile")
		})
	}, spec.Sequential())

	context("Generate called with node in the buildplan", func() {
		it.Before(func() {

			workingDir = t.TempDir()

			err := toml.NewEncoder(buf).Encode(testBuildPlan)
			fmt.Print(err)
			planPath = filepath.Join(workingDir, "plan")
			t.Setenv("CNB_BP_PLAN_PATH", planPath)

			Expect(os.WriteFile(planPath, buf.Bytes(), 0600)).To(Succeed())

			os.Chdir(workingDir)
		})

		it("Node specific version of node requested", func() {
			generateResult, err = ubi8nodeenginebuildpackextension.Generate()(packit.GenerateContext{
				WorkingDir: workingDir,
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "node",
						},
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(generateResult).NotTo(Equal(nil))
		})
	}, spec.Sequential())

}
