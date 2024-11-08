package encryption

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/docker/docker/api/types/swarm"
	"github.com/ptah-sh/ptah-agent/internal/app/ptah-agent/docker/config"

	dockerClient "github.com/docker/docker/client"
)

const encryptionKeyName = "ptah_encryption_key"

type EncryptionKeyPair struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

// GetKeyPair checks for existing key or generates a new one
func GetKeyPair(ctx context.Context, docker *dockerClient.Client) (*EncryptionKeyPair, error) {
	existingConfig, err := config.GetByName(ctx, docker, encryptionKeyName)
	if err != nil && !errors.Is(err, config.ErrConfigNotFound) {
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

	_, err = docker.ConfigCreate(ctx, swarm.ConfigSpec{
		Annotations: swarm.Annotations{
			Name:   encryptionKeyName,
			Labels: map[string]string{},
		},
		Data: keyPairJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save encryption key to Docker config: %v", err)
	}

	return keyPair, nil
}
