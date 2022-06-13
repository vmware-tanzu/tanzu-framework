// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,nolintlint
package tkgs

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

var _ = Describe("TKGS - Create Workload Cluster test cases", func() {
	var (
		options   framework.CreateClusterOptions
		logsDir string
		tkgCtlClient tkgctl.TKGClient
		clusterConfigFile string
		err error
	)
	BeforeEach(func(){
		options = framework.CreateClusterOptions{
			ClusterName:                 clusterName,
			Namespace:                   "default",
			Plan:                        "dev",
		}
		logsDir = filepath.Join(artifactsFolder, "logs")
		tkgCtlClient, err = tkgctl.New(tkgctl.Options{
			ConfigDir: e2eConfig.TkgConfigDir,
			LogOptions: tkgctl.LoggingOptions{
				File:      filepath.Join(logsDir, clusterName+".log"),
				Verbosity: e2eConfig.TkgCliLogLevel,
			},
		})
		Expect(err).To(BeNil())
	})
	// TKG-11577
	Context("Create Workload cluster with Legacy Config file", func() {
		BeforeEach(func() {
			
		})
		Context("Cluster Plan is dev", func() {
			BeforeEach(func() {
				options.Plan = "dev"

				clusterConfigFile, err = framework.GetTempClusterConfigFile(e2eConfig.TkgClusterConfigPath, &options)
				Expect(err).To(BeNil())
				err := e2eConfig.SaveWorkloadClusterOptions(clusterConfigFile)
				Expect(err).To(BeNil())
				defer os.Remove(clusterConfigFile)

				err = tkgCtlClient.ConfigCluster(tkgctl.CreateClusterOptions{
					ClusterConfigFile: clusterConfigFile,
					Edition:           "tkg",
				})
				Expect(err).To(BeNil())
				if err!= nil{
					fmt.Println(err)
				}
			
			})
			When("When cluster class cli feature flag (features.global.package-based-lcm-beta) set true",func() {
				BeforeEach(func() {
					//set the flag as true -  (features.global.package-based-lcm-beta)
				})
				It("", func(){
					err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
						ClusterConfigFile: clusterConfigFile,
						Edition:           "tkg",
					})
					Expect(err).To(BeNil())
				})
			})
			/*
			When("When cluster class cli feature flag (features.global.package-based-lcm-beta) set false",func() {
				BeforeEach(func() {
					//set the flag as false -  (features.global.package-based-lcm-beta)
				})
				It("", func(){
					err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
						ClusterConfigFile: clusterConfigFile,
						Edition:           "tkg",
					})
					Expect(err).To(BeNil())
				})
			})
			*/
		})
		/*
		Context("Cluster Plan is prod", func() {
			BeforeEach(func() {
				options.Plan = "prod"
				clusterConfigFile, err = framework.GetTempClusterConfigFile(e2eConfig.TkgClusterConfigPath, &options)
				Expect(err).To(BeNil())
				defer os.Remove(clusterConfigFile)

				err = tkgCtlClient.ConfigCluster(tkgctl.CreateClusterOptions{
					ClusterConfigFile: clusterConfigFile,
					Edition:           "tkg",
				})
				Expect(err).To(BeNil())
				
			})
			When("When cluster class cli feature flag (features.global.package-based-lcm-beta) set true",func() {
				BeforeEach(func() {
					//set the flag as true -  (features.global.package-based-lcm-beta)
				})
				It("", func(){
					err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
						ClusterConfigFile: clusterConfigFile,
						Edition:           "tkg",
					})
					Expect(err).To(BeNil())
				})
			})
	
			When("When cluster class cli feature flag (features.global.package-based-lcm-beta) set false",func() {
				BeforeEach(func() {
					//set the flag as false -  (features.global.package-based-lcm-beta)
				})
				It("", func(){
					err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
						ClusterConfigFile: clusterConfigFile,
						Edition:           "tkg",
					})
					Expect(err).To(BeNil())
				})
			})
		})*/
	})
	/*
	// TKG-11579
	Context("Create Workload cluster with Legacy Config file", func() {
		BeforeEach(func() {
			options = framework.CreateClusterOptions{
				ClusterName:                 clusterName,
				Namespace:                   "default",
				Plan:                        "dev",
			}
		})
		Context("Cluster Plan is dev", func() {
			BeforeEach(func() {
				options.Plan = "dev"

				clusterConfigFile, err = framework.GetTempClusterConfigFile(e2eConfig.TkgClusterConfigPath, &options)
				Expect(err).To(BeNil())
				defer os.Remove(clusterConfigFile)

				err = tkgCtlClient.ConfigCluster(tkgctl.CreateClusterOptions{
					ClusterConfigFile: clusterConfigFile,
					Edition:           "tkg",
				})
				Expect(err).To(BeNil())
			
			})
			When("When cluster class cli feature flag (features.global.package-based-lcm-beta) set true",func() {
				BeforeEach(func() {
					//set the flag as true -  (features.global.package-based-lcm-beta)
				})
				It("", func(){
					err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
						ClusterConfigFile: clusterConfigFile,
						Edition:           "tkg",
					})
					Expect(err).To(BeNil())
				})
			})
	
			When("When cluster class cli feature flag (features.global.package-based-lcm-beta) set false",func() {
				BeforeEach(func() {
					//set the flag as false -  (features.global.package-based-lcm-beta)
				})
				It("", func(){
					err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
						ClusterConfigFile: clusterConfigFile,
						Edition:           "tkg",
					})
					Expect(err).To(BeNil())
				})
			})
		})
		Context("Cluster Plan is prod", func() {
			BeforeEach(func() {
				options.Plan = "prod"
				clusterConfigFile, err = framework.GetTempClusterConfigFile(e2eConfig.TkgClusterConfigPath, &options)
				Expect(err).To(BeNil())
				defer os.Remove(clusterConfigFile)

				err = tkgCtlClient.ConfigCluster(tkgctl.CreateClusterOptions{
					ClusterConfigFile: clusterConfigFile,
					Edition:           "tkg",
				})
				Expect(err).To(BeNil())
				
			})
			When("When cluster class cli feature flag (features.global.package-based-lcm-beta) set true",func() {
				BeforeEach(func() {
					//set the flag as true -  (features.global.package-based-lcm-beta)
				})
				It("", func(){
					err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
						ClusterConfigFile: clusterConfigFile,
						Edition:           "tkg",
					})
					Expect(err).To(BeNil())
				})
			})
	
			When("When cluster class cli feature flag (features.global.package-based-lcm-beta) set false",func() {
				BeforeEach(func() {
					//set the flag as false -  (features.global.package-based-lcm-beta)
				})
				It("", func(){
					err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
						ClusterConfigFile: clusterConfigFile,
						Edition:           "tkg",
					})
					Expect(err).To(BeNil())
				})
			})
		})
	})

	// TKG-11578
	Context("Create Workload cluster with ClusterClass.YAML file", func() {
		Context("Legacy Workload cluster creation", func() {
			When("When cluster class cli feature flag (features.global.package-based-lcm-beta) set true",func() {
		
			})
			When("When cluster class cli feature flag (features.global.package-based-lcm-beta) set false",func() {
	
			})
		})
	})
	// TKG-11261
	Context("Create Workload cluster with ClusterClass.YAML in dry-run mode", func() {
		Context("Cluster Plan is dev", func() {
			When("When cluster class cli feature flag (features.global.package-based-lcm-beta) set true",func() {
	
			})
	
			When("When cluster class cli feature flag (features.global.package-based-lcm-beta) set false",func() {
	
			})
		})
		Context("Cluster Plan is prod", func() {
			When("When cluster class cli feature flag (features.global.package-based-lcm-beta) set true",func() {
	
			})
	
			When("When cluster class cli feature flag (features.global.package-based-lcm-beta) set false",func() {
	
			})
		})
	})
	*/
})
