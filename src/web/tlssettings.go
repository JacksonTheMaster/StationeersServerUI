package web

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/api"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/setup/update"
)

const maxTLSUploadSize = 4 << 20

func normalizeCertificate(data []byte) ([]byte, error) {
	var certificates []byte
	rest := data
	for {
		block, remaining := pem.Decode(rest)
		if block == nil {
			break
		}
		rest = remaining
		if block.Type != "CERTIFICATE" {
			continue
		}
		if _, err := x509.ParseCertificate(block.Bytes); err != nil {
			return nil, fmt.Errorf("invalid X.509 certificate: %w", err)
		}
		certificates = append(certificates, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: block.Bytes})...)
	}
	if len(certificates) > 0 {
		return certificates, nil
	}
	if _, err := x509.ParseCertificate(data); err != nil {
		return nil, fmt.Errorf("certificate must be PEM or DER X.509 data")
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: data}), nil
}

func normalizePrivateKey(data []byte) ([]byte, error) {
	der := data
	if block, _ := pem.Decode(data); block != nil {
		if x509.IsEncryptedPEMBlock(block) {
			return nil, fmt.Errorf("encrypted private keys are not supported")
		}
		der = block.Bytes
	}

	var key any
	var err error
	if key, err = x509.ParsePKCS8PrivateKey(der); err != nil {
		if key, err = x509.ParsePKCS1PrivateKey(der); err != nil {
			if key, err = x509.ParseECPrivateKey(der); err != nil {
				return nil, fmt.Errorf("private key must be unencrypted PKCS#8, RSA PKCS#1, or EC SEC1 PEM/DER data")
			}
		}
	}
	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unsupported private key: %w", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8}), nil
}

func replaceTLSFiles(certPEM, keyPEM []byte) error {
	certPath, keyPath := config.GetTLSCertPath(), config.GetTLSKeyPath()
	if filepath.Dir(certPath) != filepath.Dir(keyPath) {
		return fmt.Errorf("certificate and key paths must share a directory")
	}
	if err := os.MkdirAll(filepath.Dir(certPath), 0755); err != nil {
		return err
	}
	oldCert, certReadErr := os.ReadFile(certPath)
	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		return err
	}
	if err := os.Chmod(certPath, 0644); err != nil {
		return err
	}
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		if certReadErr == nil {
			_ = os.WriteFile(certPath, oldCert, 0644)
		}
		return err
	}
	if err := os.Chmod(keyPath, 0600); err != nil {
		return err
	}
	return nil
}

func SaveTLSCertificateHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxTLSUploadSize)
	if err := r.ParseMultipartForm(maxTLSUploadSize); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_certificate_upload", "Certificate upload is too large or invalid")
		return
	}
	readUpload := func(name string) ([]byte, error) {
		file, _, err := r.FormFile(name)
		if err != nil {
			return nil, fmt.Errorf("select a %s file", name)
		}
		defer file.Close()
		return io.ReadAll(io.LimitReader(file, maxTLSUploadSize+1))
	}
	certData, err := readUpload("certificate")
	if err != nil {
		writeTLSError(w, err)
		return
	}
	keyData, err := readUpload("privateKey")
	if err != nil {
		writeTLSError(w, err)
		return
	}
	certPEM, err := normalizeCertificate(certData)
	if err != nil {
		writeTLSError(w, err)
		return
	}
	keyPEM, err := normalizePrivateKey(keyData)
	if err != nil {
		writeTLSError(w, err)
		return
	}
	if _, err := tls.X509KeyPair(certPEM, keyPEM); err != nil {
		writeTLSError(w, fmt.Errorf("certificate and private key do not form a valid matching pair: %w", err))
		return
	}
	if err := replaceTLSFiles(certPEM, keyPEM); err != nil {
		api.WriteError(w, http.StatusInternalServerError, "certificate_save_failed", "Failed to save TLS files")
		return
	}
	api.WriteData(w, http.StatusAccepted, api.RestartReceipt{
		State: "restarting", Message: "TLS certificate saved. SSUI is restarting",
	})
	go func() {
		time.Sleep(750 * time.Millisecond)
		logger.Security.Info("TLS certificate changed; restarting SSUI")
		update.RestartMySelf()
	}()
}

func writeTLSError(w http.ResponseWriter, err error) {
	api.WriteError(w, http.StatusUnprocessableEntity, "invalid_certificate", err.Error())
}
