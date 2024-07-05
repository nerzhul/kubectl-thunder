package kubernetes

import (
	"os"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getRestConfig() (*rest.Config, error) {
	if kubeConfig := os.Getenv("KUBECONFIG"); kubeConfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeConfig)
	}

	return rest.InClusterConfig()
}

func CreateClient() (*kubernetes.Clientset, error) {
	config, err := getRestConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func CreateDynamicClient() (*dynamic.DynamicClient, error) {
	config, err := getRestConfig()
	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(config)
}
