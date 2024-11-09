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

	"golang.org/x/crypto/ssh"

	"github.com/docker/docker/api/types/swarm"
	dockerClient "github.com/docker/docker/client"
	"github.com/ptah-sh/ptah-agent/internal/app/ptah-agent/docker/config"
)

const sshKeyName = "ptah_ssh_key"

type SshKeyPair struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

// GetSshKeyPair generates a new RSA keypair for SSH,
// stores it in Docker Swarm configs
// Returns the public key in authorized_keys format
func GetSshKeyPair(ctx context.Context, docker *dockerClient.Client) (*SshKeyPair, error) {
	existingConfig, err := config.GetByName(ctx, docker, sshKeyName)
	if err != nil && !errors.Is(err, config.ErrConfigNotFound) {
		return nil, fmt.Errorf("failed to check for existing SSH key: %w", err)
	}

	if existingConfig != nil {
		var keyPair SshKeyPair
		if err := json.Unmarshal(existingConfig.Spec.Data, &keyPair); err != nil {
			return nil, fmt.Errorf("failed to unmarshal existing SSH key pair: %w", err)
		}
		return &keyPair, nil
	}

	// Generate new key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Convert private key to PEM format
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyBytes := pem.EncodeToMemory(privateKeyPEM)

	// Generate public key in SSH authorized_keys format
	sshPublicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SSH public key: %w", err)
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(sshPublicKey)

	sshKeyPair := &SshKeyPair{
		PrivateKey: string(privateKeyBytes),
		PublicKey:  string(publicKeyBytes),
	}

	// Convert SSH key pair to JSON bytes
	keyPairJSON, err := json.Marshal(sshKeyPair)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SSH key pair: %w", err)
	}

	// Store key pair in Docker config
	_, err = docker.ConfigCreate(ctx, swarm.ConfigSpec{
		Annotations: swarm.Annotations{
			Name:   sshKeyName,
			Labels: map[string]string{},
		},
		Data: keyPairJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

	return sshKeyPair, nil
}
