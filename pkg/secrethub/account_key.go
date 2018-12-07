package secrethub

import (
	"github.com/keylockerbv/secrethub-go/pkg/api"
	"github.com/keylockerbv/secrethub-go/pkg/crypto"
)

// DefaultAccountKeyLength defines the default bit size for account keys.
const DefaultAccountKeyLength = 4096

func generateAccountKey() (*crypto.RSAKey, error) {
	return crypto.GenerateRSAKey(DefaultAccountKeyLength)
}

// AccountKeyService handles operations on SecretHub account keys.
type AccountKeyService interface {
	// Create creates an account key for the client's credential.
	Create() (*api.EncryptedAccountKey, error)
	// Exists returns whether an account key exists for the client's credential.
	Exists() (bool, error)
}

type accountKeyService struct {
	client *Client
}

// newAccountKeyService creates a new accountKeyService
func newAccountKeyService(client *Client) accountKeyService {
	return accountKeyService{
		client: client,
	}
}

// Create creates an account key for the clients credential.
func (s accountKeyService) Create() (*api.EncryptedAccountKey, error) {
	key, err := generateAccountKey()
	if err != nil {
		return nil, err
	}
	return s.client.CreateAccountKey(key)
}

// Exists returns whether an account key exists for the client's credential.
func (s accountKeyService) Exists() (bool, error) {
	_, err := s.client.getAccountKey()
	if err == api.ErrAccountKeyNotFound || err == api.ErrCredentialNotKeyed {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
