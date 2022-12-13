package utils

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	KubeConfigSecretName = "admin-kubeconfig"
	KubeConfigSecretKey  = "kubeconfig"
)

// GetHostedKubeConfig get infra cluster client and HostedCluster name
// The function gets the secret contains the kubeconfig of the HostedCluster from the
// infra cluster and returns it
func GetHostedKubeConfig(c client.Client, hostedclustername string) ([]byte, error) {
	kubeconfig := &corev1.Secret{}
	secretNamespacedName := types.NamespacedName{
		Namespace: "clusters-" + hostedclustername,
		Name:      KubeConfigSecretName,
	}
	if err := c.Get(context.Background(), secretNamespacedName, kubeconfig); err != nil {
		return nil, err
	}
	return kubeconfig.Data[KubeConfigSecretKey], nil
}

// GetHostedKubeRestConfig get infra cluster client and HostedCluster name and creates
// clientConfig from the HostedCluster kubeconfig and returns it
func GetHostedKubeRestConfig(c client.Client, hostedclustername string) (*rest.Config, error) {
	config, err := GetHostedKubeConfig(c, hostedclustername)
	if err != nil {
		return nil, err
	}
	clientConfig, err := clientcmd.NewClientConfigFromBytes(config)
	if err != nil {
		return nil, err
	}
	return clientConfig.ClientConfig()
}
