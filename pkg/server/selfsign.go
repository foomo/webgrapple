package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/foomo/webgrapple/pkg/log"
	"github.com/pkg/errors"
)

// derived from : https://raw.githubusercontent.com/golang/go/master/src/crypto/tls/generate_cert.go
func selfsign(l log.Logger, hosts []string, certFile, keyFile string) error {
	const timeFormat = "Jan 2 15:04:05 2006"
	var (
		validFrom = time.Now().Format(timeFormat)
		validFor  = 365 * 24 * time.Hour
		isCA      = false
	)

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return errors.Wrap(err, "failed to generate private key")
	}

	var notBefore time.Time
	if len(validFrom) == 0 {
		notBefore = time.Now()
	} else {
		notBefore, err = time.Parse(timeFormat, validFrom)
		if err != nil {
			return errors.Wrap(err, "failed to parse creation date")
		}
	}

	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return errors.Wrap(err, "failed to generate serial number")
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range hosts {
		// if h == "" {
		// maybe check for /etc/hosts
		// log.Fatalf("Missing required --host parameter")
		// }
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return errors.Wrap(err, "failed to create certificate")
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return errors.Wrap(err, "failed to open cert.pem for writing")
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return errors.Wrap(err, "failed to write data to cert.pem")
	}
	if err := certOut.Close(); err != nil {
		return errors.Wrap(err, "error closing cert.pem")
	}

	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrap(err, "failed to open key.pem for writing")
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return errors.Wrap(err, "unable to marshal private key")
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return errors.Wrap(err, "failed to write data to key.pem")
	}
	if err := keyOut.Close(); err != nil {
		return errors.Wrap(err, "error closing key.pem")
	}
	l.Info("made certs")
	return nil
}
