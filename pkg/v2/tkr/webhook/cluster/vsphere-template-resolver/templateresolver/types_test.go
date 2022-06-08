package templateresolver

import (
	. "github.com/onsi/ginkgo"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"

	. "github.com/onsi/gomega"
)

var _ = Describe("Test all types", func() {
	Context("Result.String()", func() {
		When("result type contains information", func() {
			It("should print all data in Result", func() {
				t := TemplateQuery{
					OVAVersion: "ovaVersion",
					OSInfo: v1alpha3.OSInfo{
						Name: "osName",
					},
				}
				r := Result{
					ControlPlane: &OVATemplateResult{
						t: &TemplateResult{
							TemplatePath: "path",
							TemplateMOID: "moid",
						},
					},
					MachineDeployments: &OVATemplateResult{},
					UsefulErrorMessage: "usefulMessage",
				}
				Expect(r.String()).To(ContainSubstring("{controlPlane: {OVATemplateResult: 'map[{ovaVersion { osName  }}:{TemplatePath: 'path', TemplateMOID: 'moid'}]'}, machineDeployments: {OVATemplateResult: 'map[]'} usefulErrorMessage:'usefulMessage'}"))
			})
		})
	})
	Context("Query.String()", func() {
		When("query is not empty", func() {
			It("should print all data in Query", func() {
				q := Query{
					ControlPlane: map[TemplateQuery]struct{}{
						{OVAVersion: "cpQuery"}: {},
					},
					MachineDeployments: map[TemplateQuery]struct{}{
						{OVAVersion: "mdQuery"}: {},
					},
				}
				Expect(q.String()).To(ContainSubstring("{controlPlane: map[{cpQuery {   }}:{}], machineDeployments: map[{mdQuery {   }}:{}]}"))
			})
		})
	})
	Context("TemplateQuery.String()", func() {
		When("TemplateQuery is not empty", func() {
			It("should print all data in TemplateQuery", func() {
				tq := TemplateQuery{
					OVAVersion: "ovaVersion",
					OSInfo: v1alpha3.OSInfo{
						Type:    "type",
						Name:    "name",
						Version: "version",
						Arch:    "arch",
					},
				}
				Expect(tq.String()).To(ContainSubstring("{OVA version: 'ovaVersion', OSInfo: '{type name version arch}'}"))
			})
		})
	})
})
