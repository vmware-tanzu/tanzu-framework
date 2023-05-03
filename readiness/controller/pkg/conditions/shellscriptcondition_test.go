package conditions

/*import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
)

var _ = Describe("Readiness controller", func() {
	It("should succeed when querying an existing namespaced resource", func() {
		newPod := v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testpod1",
				Namespace: "default",
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "test-container",
						Image: "test:tag",
					},
				},
			},
		}

		err := k8sClient.Create(context.TODO(), &newPod)
		Expect(err).To(BeNil())

		state, msg := NewShellScriptConditionFunc()(context.TODO(), &corev1alpha2.ShellScriptCondition{
			Script: `
count=$(kubectl get pods -n kube-system | grep scheduler | wc -l)
echo $count
if [ $count -gt 0 ]
then
    exit 0
else
    exit 1
fi
			`,
		})

		Expect(state).To(Equal(corev1alpha2.ConditionSuccessState))
		Expect(msg).To(Equal(""))
	})

})
*/
