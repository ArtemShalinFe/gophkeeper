//go:build usetempdir
// +build usetempdir

package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path"
	"testing"
	"time"

	"github.com/ArtemShalinFe/gophkeeper/internal/config"
)

func generateTLS(t *testing.T, bits int, сertFilePath string, privateKeyFilePath string, publicKeyFilePath string) error {
	t.Helper()

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"NoOrganization"},
			Country:      []string{"RU"},
		},
		DNSNames:     []string{"*"},
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return fmt.Errorf("an occured error when generate rsa key, err: %w", err)
	}
	publicKey := &privateKey.PublicKey
	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, publicKey, privateKey)
	if err != nil {
		return fmt.Errorf("an occured error when create cert, err: %w", err)
	}
	crtPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	err = os.WriteFile(сertFilePath, crtPEM, 0600)
	if err != nil {
		return fmt.Errorf("an occured error when write public key, err: %w", err)
	}

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("an occured error when marshal private key, err: %w", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	err = os.WriteFile(privateKeyFilePath, privateKeyPEM, 0600)
	if err != nil {
		return fmt.Errorf("an occured error when write private key, err: %w", err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return fmt.Errorf("an occured error when marshal public key, err: %w", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	err = os.WriteFile(publicKeyFilePath, publicKeyPEM, 0600)
	if err != nil {
		return fmt.Errorf("an occured error when write public key, err: %w", err)
	}

	return nil
}

func Test_serverCreds(t *testing.T) {
	tmpDir := os.TempDir()

	PrivateCryptoKeyNotFound := "PrivateCryptoKeyNotFound.pem"
	CertFilePathNotFound := "CertFilePathNotFound.cert"

	certFilePath := path.Join(tmpDir, "cert.cert")
	privKeyPath := path.Join(tmpDir, "priv.pem")
	pubKeyPath := path.Join(tmpDir, "pub.pem")

	if err := generateTLS(t, 2048, certFilePath, privKeyPath, pubKeyPath); err != nil {
		t.Errorf("an occured error when generating TLS creds, err: %v", err)
	}

	defer func() {
		if err := os.Remove(certFilePath); err != nil {
			t.Errorf("an occured error when removing certfile, err: %v", err)
		}
	}()

	defer func() {
		if err := os.Remove(privKeyPath); err != nil {
			t.Errorf("an occured error when removing privkeyfile, err: %v", err)
		}
	}()

	defer func() {
		if err := os.Remove(pubKeyPath); err != nil {
			t.Errorf("an occured error when removing pubkeyfile, err: %v", err)
		}
	}()

	tests := []struct {
		name    string
		cfg     *config.ServerCfg
		wantErr bool
	}{
		{
			name:    "check insecure creds",
			cfg:     &config.ServerCfg{},
			wantErr: false,
		},
		{
			name: "check creds from real files",
			cfg: &config.ServerCfg{
				Addr:             "",
				PrivateCryptoKey: privKeyPath,
				CertFilePath:     certFilePath,
			},
			wantErr: false,
		},
		{
			name: "check creds from random files",
			cfg: &config.ServerCfg{
				Addr:             "",
				PrivateCryptoKey: path.Join(tmpDir, PrivateCryptoKeyNotFound),
				CertFilePath:     path.Join(tmpDir, CertFilePathNotFound),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := serverCreds(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("serverCreds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
