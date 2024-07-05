package commands

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"log"
	"strings"

	k8s "github.com/nerzhul/kubectl-thunder/pkg/kubernetes"
	"github.com/urfave/cli/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Secrets_find_expiring_certificates(c *cli.Context) error {
	kClient, err := k8s.CreateClient()
	if err != nil {
		return err
	}

	reportOnlyUsed := boolArg(c, "used-only")
	reportOnlyUnused := boolArg(c, "unused-only")
	deleteUseless := boolArg(c, "delete")
	afterDays := u64Arg(c, "after")

	log.Println("listing expiring TLS secrets...")
	if deleteUseless {
		log.Println("!! DELETE FLAG IS SET, EXPIRED SECRETS WILL BE DELETED !!")
	}

	expiredCerts, err := getExpiredCertificates(kClient, afterDays)
	if err != nil {
		return err
	}

	totalCerts := 0

	for _, cert := range expiredCerts {
		if !reportOnlyUsed && !reportOnlyUnused {
			log.Printf("Certificate %s/%s has expired (date: %s)", cert.Namespace, cert.Name, cert.NotAfter)
			totalCerts++
			if deleteUseless {
				log.Printf("Deleting certificate %s/%s", cert.Namespace, cert.Name)
				err := kClient.CoreV1().Secrets(cert.Namespace).Delete(context.TODO(), cert.Name, metav1.DeleteOptions{})
				if err != nil {
					return err
				}
			}

			continue
		}

		ingresses, err := getIngressUsingSecret(kClient, cert.Namespace, cert.Name)
		if err != nil {
			return err
		}

		if len(ingresses) == 0 {
			if reportOnlyUnused {
				log.Printf("Certificate %s/%s has expired (date: %s) and is not used by any ingresses", cert.Namespace, cert.Name, cert.NotAfter)
				totalCerts++
				if deleteUseless {
					log.Printf("Removing certificate %s/%s", cert.Namespace, cert.Name)
					err := kClient.CoreV1().Secrets(cert.Namespace).Delete(context.TODO(), cert.Name, metav1.DeleteOptions{})
					if err != nil {
						return err
					}
				}
			}
			continue
		}

		ingressList := []string{}
		for _, ingress := range ingresses {
			ingressList = append(ingressList, ingress.Namespace+"/"+ingress.Name)
		}

		if !reportOnlyUnused {
			log.Printf("Certificate %s/%s has expired (date: %s) and is still used by ingresses: %v", cert.Namespace, cert.Name, cert.NotAfter, ingressList)
			totalCerts++
			if deleteUseless {
				log.Printf("Removing certificate %s/%s", cert.Namespace, cert.Name)
				err := kClient.CoreV1().Secrets(cert.Namespace).Delete(context.TODO(), cert.Name, metav1.DeleteOptions{})
				if err != nil {
					return err
				}
			}
		}
	}

	log.Printf("found %d expired certificates", totalCerts)

	return nil
}

func Secrets_find_certificates_by_san(c *cli.Context) error {
	kClient, err := k8s.CreateClient()
	if err != nil {
		return err
	}

	san := c.String("san")
	wildcardMatch := c.Bool("wildcard-match")
	wildcardSAN := ""
	if wildcardMatch {
		splSan := strings.Split(san, ".")
		if len(splSan) <= 1 {
			wildcardSAN = "*"
		} else {
			wildcardSAN = "*." + strings.Join(splSan[1:], ".")
		}
	}

	secrets, err := fetchTLSSecrets(kClient)
	if err != nil {
		return err
	}

	totalCerts := 0
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
			for _, name := range cert.DNSNames {
				if name == san {
					log.Printf("Found certificate %s/%s with SAN %s", secret.Namespace, secret.Name, san)
					totalCerts++
					break
				} else if wildcardMatch && name == wildcardSAN {
					log.Printf("Found certificate %s/%s with SAN %s (wildcard match)", secret.Namespace, secret.Name, san)
					totalCerts++
					break
				}
			}
		}
	}

	log.Printf("found %d matching certificates", totalCerts)

	return nil
}
