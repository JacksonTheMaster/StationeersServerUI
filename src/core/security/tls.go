package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
)

// EnsureTLSCerts ensures TLS certificates exist and are valid at config.GetTLSCertPath() and config.GetTLSKeyPath(), generating self-signed ones if needed.
func EnsureTLSCerts() error {
	logger.Security.Debug("=== Starting TLS certificate check ===")

	certPath := config.GetTLSCertPath()
	keyPath := config.GetTLSKeyPath()
	logger.Security.Debug(fmt.Sprintf("Cert path: %s", certPath))
	logger.Security.Debug(fmt.Sprintf("Key path: %s", keyPath))

	// Check if cert and key files exist
	certExists := fileExists(certPath)
	keyExists := fileExists(keyPath)
	logger.Security.Debug(fmt.Sprintf("Cert exists: %t, Key exists: %t", certExists, keyExists))

	tlsDir := config.GetSSUIFolder() + "tls/"

	if _, err := os.Stat(tlsDir); os.IsNotExist(err) {
		logger.Security.Debug("TLS directory doesn't exist, creating...")
		if err := os.MkdirAll(tlsDir, os.ModePerm); err != nil {
			logger.Security.Error("Failed to create directory " + tlsDir + ": " + err.Error())
			return nil
		}
	}

	if certExists && keyExists {
		logger.Security.Debug("Both cert and key files exist, checking validity...")

		// Load and check if the cert is still valid
		certData, err := os.ReadFile(certPath)
		if err != nil {
			logger.Security.Error(fmt.Sprintf("Failed to read cert file: %v", err))
			return fmt.Errorf("failed to read cert file: %v", err)
		}

		certBlock, _ := pem.Decode(certData)
		if certBlock == nil {
			logger.Security.Error("Failed to decode PEM certificate")
			return fmt.Errorf("failed to decode PEM certificate")
		}

		cert, err := x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			logger.Security.Error(fmt.Sprintf("Failed to parse certificate: %v", err))
			return fmt.Errorf("failed to parse certificate: %v", err)
		}

		// Log certificate details
		logger.Security.Debug(fmt.Sprintf("Certificate Serial: %s", cert.SerialNumber.String()))
		logger.Security.Debug(fmt.Sprintf("Certificate NotBefore: %s", cert.NotBefore.Format(time.RFC3339)))
		logger.Security.Debug(fmt.Sprintf("Certificate NotAfter: %s", cert.NotAfter.Format(time.RFC3339)))
		logger.Security.Debug(fmt.Sprintf("Current time: %s", time.Now().Format(time.RFC3339)))

		// Check if expired or near expiry (within 10 days of 90-day validity)
		now := time.Now()
		nearExpiry := now.Add(10 * 24 * time.Hour)

		if now.After(cert.NotAfter) {
			logger.Security.Warn("Certificate is expired, regenerating...")
		} else if nearExpiry.After(cert.NotAfter) {
			logger.Security.Warn("Certificate is near expiry (within 10 days), regenerating...")
		} else {
			// Cert is valid, we're done
			logger.Security.Debug("Certificate is valid and not near expiry, using existing cert")
			logger.Security.Debug("=== TLS certificate check complete - using existing ===")
			return nil
		}
	} else {
		logger.Security.Debug("Certificate or key file missing, will generate new ones")
	}

	// Generate a new self-signed cert if files are missing or expired
	logger.Security.Debug("Calling generateSelfSignedCert()...")
	if err := generateSelfSignedCert(); err != nil {
		logger.Security.Error(fmt.Sprintf("Failed to generate self-signed cert: %v", err))
		return fmt.Errorf("failed to generate self-signed cert: %v", err)
	}

	logger.Security.Info("Generated new self-signed TLS certificates at " + certPath + " and " + keyPath)
	logger.Security.Debug("=== TLS certificate check complete - generated new ===")
	return nil
}

// generateSelfSignedCert creates a self-signed certificate and key pair at config.GetTLSCertPath() and config.GetTLSKeyPath().
func generateSelfSignedCert() error {
	logger.Security.Debug("=== Generating new self-signed certificate ===")

	certPath := config.GetTLSCertPath()
	keyPath := config.GetTLSKeyPath()

	dir := filepath.Dir(certPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Generate a private key
	logger.Security.Debug("Generating RSA private key...")
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	serialNumber, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	logger.Security.Debug(fmt.Sprintf("Generated serial number: %s", serialNumber.String()))

	// Create a certificate template
	now := time.Now()
	notAfter := now.Add(90 * 24 * time.Hour)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"StationeersServerUI"},
			CommonName:   "localhost",
		},
		NotBefore:             now,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost", "ssui.local"},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}

	logger.Security.Debug(fmt.Sprintf("Certificate template - NotBefore: %s, NotAfter: %s",
		now.Format(time.RFC3339), notAfter.Format(time.RFC3339)))

	// Create the certificate
	logger.Security.Debug("Creating certificate...")
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	// Write the certificate to file
	logger.Security.Debug(fmt.Sprintf("Writing certificate to: %s", certPath))
	certFile, err := os.Create(certPath)
	if err != nil {
		return err
	}
	defer certFile.Close()
	pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// Write the private key to file
	logger.Security.Debug(fmt.Sprintf("Writing private key to: %s", keyPath))
	keyFile, err := os.Create(keyPath)
	if err != nil {
		return err
	}
	defer keyFile.Close()
	pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	logger.Security.Debug("=== Certificate generation complete ===")
	return nil
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
