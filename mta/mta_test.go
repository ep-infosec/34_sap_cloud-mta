package mta

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/SAP/cloud-mta/internal/fs"
)

var _ = Describe("Mta", func() {

	modules := []*Module{
		{
			Name: "backend",
			Type: "java.tomcat",
			Path: "java",
			BuildParams: map[string]interface{}{
				"builder": "maven",
			},
			Properties: map[string]interface{}{
				"backend_type": nil,
			},
			PropertiesMetaData: map[string]MetaData{
				"backend_type": {
					Optional:     falsePtr(),
					OverWritable: truePtr(),
					Datatype:     "str",
				},
			},
			Parameters: map[string]interface{}{
				"domain":   nil,
				"password": "asfhuwehkew efgehk",
			},
			ParametersMetaData: map[string]MetaData{
				"domain": {
					Optional:     falsePtr(),
					OverWritable: truePtr(),
				},
			},
			Includes: []Includes{
				{
					Name: "configs",
					Path: "cfg/parameters.json",
				},
			},
			Provides: []Provides{
				{
					Name:   "backend_task",
					Public: true,
					Properties: map[string]interface{}{
						"url": "${default-url}/tasks",
					},
					PropertiesMetaData: map[string]MetaData{
						"url": {
							Optional:     truePtr(),
							OverWritable: truePtr(),
						},
					},
				},
				{
					Name:   "finished_backend_tasks",
					Public: true,
					Properties: map[string]interface{}{
						"url": "${default-url}/finishedTasks",
					},
					PropertiesMetaData: map[string]MetaData{
						"url": {
							Optional:     truePtr(),
							OverWritable: falsePtr(),
						},
					},
				},
			},
			Requires: []Requires{
				{
					Name: "database",
				},
				{
					Name: "scheduler_api",
					List: "mylist",
					Properties: map[string]interface{}{
						"scheduler_url": "~{url}",
					},
					PropertiesMetaData: map[string]MetaData{
						"scheduler_url": {
							Optional: falsePtr(),
						},
					},
					Includes: []Includes{
						{
							Name: "configs",
							Path: "cfg/parameters.json",
						},
					},
				},
			},
			DeployedAfter: []string{"scheduler"},
			Hooks: []Hook{
				{
					Name:   "hook",
					Type:   "task",
					Phases: []string{"application.before-stop.live", "application.before-stop.idle"},
					Parameters: map[string]interface{}{
						"name":    "foo-task",
						"command": "sleep 5m",
					},
				},
			},
		},
		{
			Name: "scheduler",
			Type: "javascript.nodejs",
			Provides: []Provides{
				{
					Name: "scheduler_api",
					Properties: map[string]interface{}{
						"url": "${default-url}/api/v2",
					},
				},
			},
			Requires: []Requires{
				{
					Name: "backend_task",
					Properties: map[string]interface{}{
						"task_url": "~{url}",
					},
				},
			},
		},
	}

	buildersBefore := []ProjectBuilder{
		{
			Builder: "mybuilder",
		},
	}
	buildersAfter := []ProjectBuilder{
		{
			Builder: "otherbuilder",
		},
	}

	buildParams := &ProjectBuild{
		BeforeAll: buildersBefore,
		AfterAll:  buildersAfter,
	}

	schemaVersion := "3.3"
	mta := &MTA{
		SchemaVersion: &schemaVersion,
		ID:            "com.acme.scheduling",
		Version:       "1.132.1-edfsd+ewfe",
		Parameters:    map[string]interface{}{"deployer-version": ">=1.2.0"},
		BuildParams:   buildParams,
		Modules:       modules,
		Resources: []*Resource{
			{
				Name:           "database",
				Type:           "postgresql",
				ProcessedAfter: []string{"plugins"},
			},
			{
				Name:     "plugins",
				Type:     "configuration",
				Optional: true,
				Active:   falsePtr(),
				Requires: []Requires{
					{
						Name: "scheduler_api",
						Parameters: map[string]interface{}{
							"par1": "value",
						},
						Properties: map[string]interface{}{
							"prop1": "${value}-~{url}",
						},
					},
				},
				Includes: []Includes{
					{
						Name: "configs",
						Path: "cfg/security.json",
					},
					{
						Name: "creation",
						Path: "djdk.yaml",
					},
				},
				Parameters: map[string]interface{}{
					"filter": map[string]interface{}{
						"type": "com.acme.plugin",
					},
				},
				ParametersMetaData: map[string]MetaData{
					"filter": {
						Optional:     falsePtr(),
						OverWritable: falsePtr(),
					},
				},
				Properties: map[string]interface{}{
					"plugin_name": "${name}",
					"plugin_url":  "${url}/sources",
				},
				PropertiesMetaData: map[string]MetaData{
					"plugin_name": {
						Optional: truePtr(),
					},
				},
			},
		},
		ModuleTypes: []*ModuleTypes{
			{
				Name:    "java.tomcat",
				Extends: "java",
				Parameters: map[string]interface{}{
					"buildpack": nil,
					"memory":    "256M",
				},
				ParametersMetaData: map[string]MetaData{
					"buildpack": {
						Optional:     falsePtr(),
						OverWritable: nil,
					},
				},
				Properties: map[string]interface{}{
					"TARGET_RUNTIME": "tomcat",
				},
			},
		},
		ResourceTypes: []*ResourceTypes{
			{
				Name:    "postgresql",
				Extends: "managed-service",
				Parameters: map[string]interface{}{
					"service":      "postgresql",
					"service-plan": nil,
				},
				ParametersMetaData: map[string]MetaData{
					"service-plan": {
						Optional:     falsePtr(),
						OverWritable: nil,
					},
				},
			},
		},
	}
	var _ = Describe("MTA tests", func() {

		var _ = Describe("Parsing", func() {
			It("Modules parsing - sanity", func() {
				mtaFile, err := ioutil.ReadFile("./testdata/mta.yaml")
				??(err).Should(Succeed())
				// Unmarshal file
				oMta, err := Unmarshal(mtaFile)
				??(err).Should(Succeed())
				??(oMta.Modules).Should(HaveLen(2))
				??(oMta.GetModules()).Should(Equal(modules))
			})
			It("Resource parsing - active", func() {
				mtaFile, err := ioutil.ReadFile("./testdata/mtaResourceActive.yaml")
				??(err).Should(Succeed())
				// Unmarshal file
				oMta, err := Unmarshal(mtaFile)
				??(err).Should(Succeed())
				resources := oMta.GetResources()
				??(resources).Should(HaveLen(3))

				??(resources[0].Name).Should(Equal("defaultActive"))
				??(resources[0].Active).Should(BeNil())
				??(resources[1].Name).Should(Equal("activeIsFalse"))
				??(resources[1].Active).ShouldNot(BeNil())
				??(*resources[1].Active).Should(BeFalse())
				??(resources[2].Name).Should(Equal("activeIsTrue"))
				??(resources[2].Active).ShouldNot(BeNil())
				??(*resources[2].Active).Should(BeTrue())
			})
		})

		var _ = Describe("Get methods on MTA", func() {
			It("GetModules", func() {
				??(mta.GetModules()).Should(Equal(modules))
			})
			It("GetResourceByName - Sanity", func() {
				??(mta.GetResourceByName("database")).Should(Equal(mta.Resources[0]))
				??(mta.GetResourceByName("plugins")).Should(Equal(mta.Resources[1]))
			})
			It("GetResourceByName - Negative", func() {
				res := mta.GetResourceByName("")
				??(res).Should(BeNil())
			})
			It("GetResources - Sanity ", func() {
				??(mta.GetResources()).Should(Equal(mta.Resources))
			})
			It("GetModuleByName - Sanity ", func() {
				??(mta.GetModuleByName("backend")).Should(Equal(modules[0]))
				??(mta.GetModuleByName("scheduler")).Should(Equal(modules[1]))
			})
			It("GetModuleByName - Negative ", func() {
				mod, err := mta.GetModuleByName("foo")
				??(mod).Should(BeNil())
				??(err).Should(HaveOccurred())
			})
		})

		Describe("Get methods on Module", func() {
			It("GetProvidesByName - Sanity", func() {
				module := mta.Modules[0]
				??(*(module.GetProvidesByName("backend_task"))).Should(Equal(module.Provides[0]))
				??(*(module.GetProvidesByName("finished_backend_tasks"))).Should(Equal(module.Provides[1]))
				??(module.GetProvidesByName("finished")).Should(BeNil())
			})
		})
	})

	var _ = Describe("Unmarshal", func() {
		It("Sanity", func() {
			wd, err := os.Getwd()
			??(err).Should(Succeed())
			content, err := fs.ReadFile(filepath.Join(wd, "testdata", "mta.yaml"))
			??(err).Should(Succeed())
			m, err := Unmarshal(content)
			??(err).Should(Succeed())

			??(mta).Should(BeEquivalentTo(m))
		})

		It("Wrong deployed-after value", func() {
			content, err := fs.ReadFile(getTestPath("mtaWrongDeployedAfter.yaml"))
			??(err).Should(Succeed())
			_, err = Unmarshal(content)
			??(err).Should(HaveOccurred())
			??(err.Error()).Should(ContainSubstring("line 54: cannot unmarshal !!int `1` into []string"))
		})

		It("Wrong processed-after value", func() {
			content, err := fs.ReadFile(getTestPath("mtaWrongProcessedAfter.yaml"))
			??(err).Should(Succeed())
			_, err = Unmarshal(content)
			??(err).Should(HaveOccurred())
			??(err.Error()).Should(ContainSubstring("line 69: cannot unmarshal !!int `1` into []string"))
		})

		It("Correct processed-after value", func() {
			content, err := fs.ReadFile(getTestPath("mtaCorrectProcessedAfter.yaml"))
			??(err).Should(Succeed())
			_, err = Unmarshal(content)
			??(err).ShouldNot(HaveOccurred())
		})

		It("Wrong properties-metadata value", func() {
			content, err := fs.ReadFile(getTestPath("mtaWrongMetaData.yaml"))
			??(err).Should(Succeed())
			_, err = Unmarshal(content)
			??(err).Should(HaveOccurred())
			??(err.Error()).Should(ContainSubstring("line 23: cannot unmarshal !!bool `true` into mta.metadata"))
		})
	})
})
