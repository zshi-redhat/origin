package sriovnetwork

import (
	"fmt"
	"strings"
	"time"

        exutil "github.com/openshift/origin/test/extended/util"
	e2e "k8s.io/kubernetes/test/e2e/framework"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("[Area:Networking] SRIOV CNI", func() {
	defer GinkgoRecover()

	InNetworkAttachmentContext(func() {
		oc := exutil.NewCLI("sriov", exutil.KubeConfigPath())
		f1 := oc.KubeFramework()

		InSRIOVProvisionedContext(func() {
			InSRIOVDeployedContext(func() {
				It("should advertise correct SRIOV VF numbers", func() {
					time.Sleep(1 * time.Minute)
					for _, n := range resConfList.ResourceList {
						templateArgs := fmt.Sprintf(
							"'{{ index .status.allocatable \"openshift.com/%s\" }}'",
							n.ResourceName)
						out, err := oc.AsAdmin().Run("get").Args("node", n.NodeName).
							Template(templateArgs).Output()
						Expect(err).NotTo(HaveOccurred())
						Expect(out).To(Equal(fmt.Sprintf("'%s'", n.ResourceNum)))
						By(fmt.Sprintf("Node %s allocatable output: %s", n.NodeName, out))
					}

					for _, n := range resConfList.ResourceList {
						By("Creating SRIOV Test Pod")
						err := oc.AsAdmin().Run("create").
							Args("-f", fmt.Sprintf("%s/pod-%s.yaml",
							TestDataFixture, n.ResourceName)).Execute()
						Expect(err).NotTo(HaveOccurred())

						By("Waiting for testpod become ready")
						err = wait.PollImmediate(e2e.Poll,
							3*time.Minute, func() (bool, error) {
							err = CheckPodStatus(oc,
								fmt.Sprintf("testpod-%s", n.ResourceName))
							if err != nil {
								return false, nil
							}
							return true, nil
						})
						Expect(err).NotTo(HaveOccurred())
					}

					defer func() {
						for _, n := range resConfList.ResourceList {
							oc.AsAdmin().Run("delete").
								Args("-f", fmt.Sprintf("%s/pod-%s.yaml",
								TestDataFixture, n.ResourceName)).Execute()
						}
					}()
				})
			})
		})
	})
})
