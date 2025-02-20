package tests

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	secv1 "github.com/openshift/api/security/v1"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubevirt.io/ssp-operator/internal/common"
	nodelabeller "kubevirt.io/ssp-operator/internal/operands/node-labeller"
)

var _ = Describe("Node Labeller", func() {
	var (
		clusterRoleRes               testResource
		clusterRoleBindingRes        testResource
		serviceAccountRes            testResource
		securityContextConstraintRes testResource
		configMapRes                 testResource
		daemonSetRes                 testResource
	)

	BeforeEach(func() {
		expectedLabels := expectedLabelsFor("node-labeler", common.AppComponentSchedule)
		clusterRoleRes = testResource{
			Name:           nodelabeller.ClusterRoleName,
			Resource:       &rbac.ClusterRole{},
			Namespace:      "",
			ExpectedLabels: expectedLabels,
			UpdateFunc: func(role *rbac.ClusterRole) {
				role.Rules[0].Verbs = []string{"watch"}
			},
			EqualsFunc: func(old *rbac.ClusterRole, new *rbac.ClusterRole) bool {
				return reflect.DeepEqual(old.Rules, new.Rules)
			},
		}
		clusterRoleBindingRes = testResource{
			Name:           nodelabeller.ClusterRoleBindingName,
			Resource:       &rbac.ClusterRoleBinding{},
			Namespace:      "",
			ExpectedLabels: expectedLabels,
			UpdateFunc: func(roleBinding *rbac.ClusterRoleBinding) {
				roleBinding.Subjects = nil
			},
			EqualsFunc: func(old *rbac.ClusterRoleBinding, new *rbac.ClusterRoleBinding) bool {
				return reflect.DeepEqual(old.Subjects, new.Subjects)
			},
		}
		serviceAccountRes = testResource{
			Name:           nodelabeller.ServiceAccountName,
			Namespace:      strategy.GetNamespace(),
			Resource:       &core.ServiceAccount{},
			ExpectedLabels: expectedLabels,
		}
		securityContextConstraintRes = testResource{
			Name:           nodelabeller.SecurityContextName,
			Namespace:      "",
			Resource:       &secv1.SecurityContextConstraints{},
			ExpectedLabels: expectedLabels,
			UpdateFunc: func(scc *secv1.SecurityContextConstraints) {
				scc.Users = []string{"test-user"}
			},
			EqualsFunc: func(old *secv1.SecurityContextConstraints, new *secv1.SecurityContextConstraints) bool {
				return reflect.DeepEqual(old.Users, new.Users)
			},
		}
		configMapRes = testResource{
			Name:           nodelabeller.ConfigMapName,
			Namespace:      strategy.GetNamespace(),
			Resource:       &core.ConfigMap{},
			ExpectedLabels: expectedLabels,
			UpdateFunc: func(configMap *core.ConfigMap) {
				configMap.Data = map[string]string{
					"cpu-plugin-configmap.yaml": "change data",
				}
			},
			EqualsFunc: func(old *core.ConfigMap, new *core.ConfigMap) bool {
				return reflect.DeepEqual(old.Data, new.Data)
			},
		}
		daemonSetRes = testResource{
			Name:           nodelabeller.DaemonSetName,
			Namespace:      strategy.GetNamespace(),
			Resource:       &apps.DaemonSet{},
			ExpectedLabels: expectedLabels,
			UpdateFunc: func(daemonSet *apps.DaemonSet) {
				daemonSet.Spec.Template.Spec.ServiceAccountName = "test-account"
			},
			EqualsFunc: func(old *apps.DaemonSet, new *apps.DaemonSet) bool {
				return reflect.DeepEqual(old.Spec, new.Spec)
			},
		}

		waitUntilDeployed()
	})

	Context("resource deletion", func() {
		DescribeTable("deleted cluster resource", func(res *testResource) {
			resource := res.NewResource()
			err := apiClient.Get(ctx, res.GetKey(), resource)
			Expect(err).To(HaveOccurred())
			Expect(errors.ReasonForError(err)).To(Equal(metav1.StatusReasonNotFound))
		},
			Entry("[test_id:6068]cluster role", &clusterRoleRes),
			Entry("[test_id:6067]cluster role binding", &clusterRoleBindingRes),
			Entry("[test_id:6066]security context constraint", &securityContextConstraintRes),
		)

		DescribeTable("deleted namespaced resource", func(res *testResource) {
			err := apiClient.Get(ctx, res.GetKey(), res.NewResource())
			Expect(err).To(HaveOccurred())
			Expect(errors.ReasonForError(err)).To(Equal(metav1.StatusReasonNotFound))
		},
			Entry("[test_id:6063]service account", &serviceAccountRes),
			Entry("[test_id:6064]configMap", &configMapRes),
			Entry("[test_id:6065]daemonSet", &daemonSetRes),
		)
	})

})
