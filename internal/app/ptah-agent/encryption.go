package ptah_agent

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"github.com/pkg/errors"

	"github.com/docker/docker/api/types/swarm"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

type EncryptionKeyPair struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

// getEncryptionKey checks for existing key or generates a new one
func (e *taskExecutor) getEncryptionKey(ctx context.Context) (*EncryptionKeyPair, error) {
	existingConfig, err := e.getConfigByName(ctx, "ptah_encryption_key")
	if err != nil && !errors.Is(err, ErrConfigNotFound) {
		return nil, fmt.Errorf("failed to check for existing encryption key: %v", err)
	}

	if existingConfig != nil {
		var keyPair EncryptionKeyPair

		err = json.Unmarshal(existingConfig.Spec.Data, &keyPair)

		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal existing encryption key: %v", err)
		}

		return &keyPair, nil
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key pair: %v", err)
	}

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyStr := string(pem.EncodeToMemory(privateKeyPEM))

	publicKey, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %v", err)
	}

	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKey,
	}
	publicKeyStr := string(pem.EncodeToMemory(publicKeyPEM))

	keyPair := &EncryptionKeyPair{
		PrivateKey: privateKeyStr,
		PublicKey:  publicKeyStr,
	}

	keyPairJSON, err := json.Marshal(keyPair)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal encryption key pair: %v", err)
	}

	_, err = e.createDockerConfig(ctx, &t.CreateConfigReq{
		SwarmConfigSpec: swarm.ConfigSpec{
			Annotations: swarm.Annotations{
				Name:   "ptah_encryption_key",
				Labels: map[string]string{},
			},
			Data: keyPairJSON,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save encryption key to Docker config: %v", err)
	}

	return keyPair, nil
}

func (e *taskExecutor) decryptValue(ctx context.Context, encryptedValue string) (string, error) {
	keyPair, err := e.getEncryptionKey(ctx)
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
	keyPair, err := e.getEncryptionKey(ctx)
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
