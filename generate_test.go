package ubi8nodeenginebuildpackextension_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ubi8nodeenginebuildpackextension "github.com/nodeshift/ubi8-node-engine-buildack-extension"
	"github.com/nodeshift/ubi8-node-engine-buildack-extension/fakes"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/packit/v2/cargo"

	"github.com/BurntSushi/toml"
	postal "github.com/paketo-buildpacks/packit/v2/postal"
)

type extensionTomlProps struct {
	NODEJS_VERSION string
}

func testGenerate(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect            = NewWithT(t).Expect
		workingDir        string
		planPath          string
		testBuildPlan     packit.BuildpackPlan
		buf               = new(bytes.Buffer)
		generateResult    packit.GenerateResult
		err               error
		cnbDir            string
		dependencyManager *fakes.DependencyManager
	)

	context("Generate called with NO node in buildplan", func() {
		it.Before(func() {

			workingDir = t.TempDir()
			Expect(err).NotTo(HaveOccurred())

			err := toml.NewEncoder(buf).Encode(testBuildPlan)
			fmt.Print(err)

			Expect(os.WriteFile(filepath.Join(workingDir, "plan"), buf.Bytes(), 0600)).To(Succeed())

			os.Chdir(workingDir)
		})

		it("Node no longer requested in buildplan", func() {
			dependencyManager = &fakes.DependencyManager{}
			dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{Name: "Node Engine", ID: "node", Version: "16.5.1"}

			generateResult, err = ubi8nodeenginebuildpackextension.Generate(dependencyManager)(packit.GenerateContext{
				WorkingDir: workingDir,
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
			})
			Expect(err).To(HaveOccurred())
			Expect(generateResult.BuildDockerfile).To(BeNil())
		})
	}, spec.Sequential())

	context("Generate called with node in the buildplan", func() {
		it.Before(func() {

			workingDir = t.TempDir()
			cnbDir, err = os.MkdirTemp("", "cnb")

			err := toml.NewEncoder(buf).Encode(testBuildPlan)
			fmt.Print(err)
			planPath = filepath.Join(workingDir, "plan")
			t.Setenv("CNB_BP_PLAN_PATH", planPath)

			Expect(os.WriteFile(planPath, buf.Bytes(), 0600)).To(Succeed())

			os.Chdir(workingDir)
		})

		it("Node specific version of node requested", func() {
			dependencyManager = &fakes.DependencyManager{}
			dependencyManager.ResolveCall.Returns.Dependency =
				postal.Dependency{Name: "Node Engine", ID: "node", Version: "16.5.1", Source: "172.17.0.1:5000/ubi8-paketo-run-nodejs-16"}
			generateResult, err = ubi8nodeenginebuildpackextension.Generate(dependencyManager)(packit.GenerateContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
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

			buf := new(strings.Builder)
			_, _ = io.Copy(buf, generateResult.RunDockerfile)
			Expect(buf.String()).To(Equal("FROM 172.17.0.1:5000/ubi8-paketo-run-nodejs-16"))
		})

		it("Node specific version of node requested", func() {

			extensionToml, _ := readExtensionTomlTemplateFile()

			cnbDir, err = os.MkdirTemp("", "cnb")
			os.WriteFile(cnbDir+"/extension.toml", []byte(extensionToml), 0600)

			dependencyManager := postal.NewService(cargo.NewTransport())

			generateResult, err = ubi8nodeenginebuildpackextension.Generate(dependencyManager)(packit.GenerateContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "node",
							Metadata: map[string]interface{}{
								"version":        "16",
								"version-source": "BP_NODE_VERSION",
							},
						},
					},
				},
				Stack: "ubi8-paketo",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(generateResult).NotTo(Equal(nil))

			buf := new(strings.Builder)
			_, _ = io.Copy(buf, generateResult.RunDockerfile)
			Expect(buf.String()).To(Equal("FROM 172.17.0.1:5000/ubi8-paketo-run-nodejs-16"))


		})

	}, spec.Sequential())

}

func readExtensionTomlTemplateFile() (string, error) {
	return `api = "0.7"

	[extension]
	id = "redhat-runtimes/nodejs"
	name = "RedHat Runtimes Node.js Dependency Extension"
	version = "0.0.1"
	description = "This extension installs the appropriate nodejs runtime via dnf"
	
	[metadata]
	  [metadata.default-versions]
		node = "18.*.*"
	
	  [[metadata.dependencies]]
		id = "node"
		name = "Ubi Node Extension"
		stacks = ["ubi8-paketo"]
		source = "172.17.0.1:5000/ubi8-paketo-run-nodejs-18"
		version = "18.1000"

	  [[metadata.dependencies]]
		id = "node"
		name = "Ubi Node Extension"
		stacks = ["ubi8-paketo"]
		source = "172.17.0.1:5000/ubi8-paketo-run-nodejs-16"
		version = "16.1000"
		`, nil
}
