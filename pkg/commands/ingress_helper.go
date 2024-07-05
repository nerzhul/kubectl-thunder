package commands

import (
	"context"

	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func getIngressUsingSecret(client *kubernetes.Clientset, namespace string, secretName string) ([]netv1.Ingress, error) {
	ingresses, err := client.NetworkingV1().Ingresses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	foundIngresses := []netv1.Ingress{}

	for _, ingress := range ingresses.Items {
		for _, tls := range ingress.Spec.TLS {
			if tls.SecretName == secretName {
				foundIngresses = append(foundIngresses, ingress)
			}
		}
	}

	return foundIngresses, nil
}
