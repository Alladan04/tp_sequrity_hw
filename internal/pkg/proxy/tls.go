package proxy

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"time"
)

func CreateCert(dnsNames []string, parent *x509.Certificate, parentKey crypto.PrivateKey, hoursValid int) (cert []byte, priv []byte) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Key generation failed: %v", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("Serial number generation failed: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"_"},
		},
		DNSNames:  dnsNames,
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Duration(hoursValid) * time.Hour),

		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, parent, &privateKey.PublicKey, parentKey)
	if err != nil {
		log.Fatalf("Create cert failed: %v", err)
	}
	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if pemCert == nil {
		log.Fatal("encod cert to PEM failed")
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		log.Fatalf("private key marshalling failed: %v", err)
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if pemCert == nil {
		log.Fatal("failed to encode key to PEM")
	}

	return pemCert, pemKey
}

func LoadKeyPair(host string) (cert *x509.Certificate, key any, err error) {
	cf, err := os.ReadFile("./certs/hck.crt")
	if err != nil {
		return nil, nil, err
	}

	kf, err := os.ReadFile("./certs/cert.key")
	if err != nil {
		return nil, nil, err
	}
	certBlock, _ := pem.Decode(cf)
	cert, err = x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	keyBlock, _ := pem.Decode(kf)
	key, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}
