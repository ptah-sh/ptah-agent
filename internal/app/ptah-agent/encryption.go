package ptah_agent

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"

	"github.com/pkg/errors"
	"github.com/ptah-sh/ptah-agent/internal/app/ptah-agent/encryption"
)

func (e *taskExecutor) decryptValue(ctx context.Context, encryptedValue string) (string, error) {
	keyPair, err := encryption.GetKeyPair(ctx, e.docker)
	if err != nil {
		return "", errors.Wrap(err, "failed to get encryption key")
	}

	block, _ := pem.Decode([]byte(keyPair.PrivateKey))
	if block == nil {
		return "", errors.New("failed to parse PEM block containing the private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse private key")
	}

	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedValue)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode encrypted value")
	}

	decryptedBytes, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedBytes, []byte(""))
	if err != nil {
		return "", errors.Wrap(err, "failed to decrypt value")
	}

	return string(decryptedBytes), nil
}

func (e *taskExecutor) encryptValue(ctx context.Context, value string) (string, error) {
	keyPair, err := encryption.GetKeyPair(ctx, e.docker)
	if err != nil {
		return "", errors.Wrap(err, "failed to get encryption key")
	}

	block, _ := pem.Decode([]byte(keyPair.PublicKey))
	if block == nil {
		return "", errors.New("failed to parse PEM block containing the public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse public key")
	}

	encryptedBytes, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey.(*rsa.PublicKey), []byte(value), []byte(""))
	if err != nil {
		return "", errors.Wrap(err, "failed to encrypt value")
	}

	return base64.StdEncoding.EncodeToString(encryptedBytes), nil
}
