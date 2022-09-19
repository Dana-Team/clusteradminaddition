package testUtils

import (
	"github.com/openshift/hypershift/api/v1alpha1"
	v1 "k8s.io/api/rbac/v1"
	v1api "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetHostedClusterObject(name string) *v1alpha1.HostedCluster {
	hostedCluster := &v1alpha1.HostedCluster{
		TypeMeta: v1api.TypeMeta{
			Kind:       "HostedCluster",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: v1api.ObjectMeta{
			Name: name,
		},
		Spec:   v1alpha1.HostedClusterSpec{},
		Status: v1alpha1.HostedClusterStatus{},
	}
	return hostedCluster
}

func GetClusterRoleBinding(name string) v1.ClusterRoleBinding {
	clusterRoleBinding := v1.ClusterRoleBinding{
		TypeMeta: v1api.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "v1",
		},
		ObjectMeta: v1api.ObjectMeta{
			Name: name,
		},
		Subjects: []v1.Subject{
			{Kind: "User",
				Name: name},
		},
		RoleRef: v1.RoleRef{
			Kind: "ClusterRole",
			Name: "cluster-admin",
		},
	}
	return clusterRoleBinding
}
