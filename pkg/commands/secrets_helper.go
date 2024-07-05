package commands

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type expiredSecretMetadata struct {
	Namespace string
	Name      string
	SNI       []string
	NotAfter  time.Time
}

func fetchTLSSecrets(client *kubernetes.Clientset) ([]v1.Secret, error) {
	namespaces, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	secrets := []v1.Secret{}

	for _, namespace := range namespaces.Items {
		secretsList, err := client.CoreV1().Secrets(namespace.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		for _, secret := range secretsList.Items {
			// Check if the secret is a tls secret
			if secret.Type != "kubernetes.io/tls" {
				continue
			}

			// Check if the secret contains the tls.crt key
			if _, ok := secret.Data["tls.crt"]; !ok {
				continue
			}

			secrets = append(secrets, secret)
		}
	}

	return secrets, nil
}

func getExpiredCertificates(client *kubernetes.Clientset, afterDays uint64) ([]expiredSecretMetadata, error) {
	secrets, err := fetchTLSSecrets(client)
	if err != nil {
		return nil, err
	}

	expiredCerts := []expiredSecretMetadata{}
	for _, secret := range secrets {
		certEncoded := secret.Data["tls.crt"]
		certs := []*x509.Certificate{}
		for block, rest := pem.Decode(certEncoded); block != nil; block, rest = pem.Decode(rest) {
			if block.Type == "CERTIFICATE" {
				certEncoded = block.Bytes
				cert, err := x509.ParseCertificate(certEncoded)
				if err != nil {
					break
				}

				certs = append(certs, cert)
			}

			if len(rest) == 0 {
				break
			}
		}

		// decode the certificate
		for _, cert := range certs {
			expiringDate := time.Now().Add(time.Duration(afterDays) * 24 * time.Hour)
			if expiringDate.After(cert.NotAfter) {
				expiredCerts = append(expiredCerts, expiredSecretMetadata{
					Namespace: secret.Namespace,
					Name:      secret.Name,
					SNI:       cert.DNSNames,
					NotAfter:  cert.NotAfter,
				})
			}
		}
	}

	return expiredCerts, nil
}
