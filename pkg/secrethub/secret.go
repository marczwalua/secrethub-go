package secrethub

import (
	"github.com/secrethub/secrethub-go/internals/api"
	"github.com/secrethub/secrethub-go/internals/errio"
)

// SecretService handles operations on secrets from SecretHub.
type SecretService interface {
	// Delete removes the secret at the given path.
	Delete(path string) error
	// Exists returns whether a secret exists on the given path.
	Exists(path string) (bool, error)
	// Get retrieves a Secret.
	Get(path string) (*api.Secret, error)
	// ListEvents retrieves all audit events for a given secret.
	ListEvents(path string, subjectTypes api.AuditSubjectTypeList) ([]*api.Audit, error)

	// Versions returns a SecretVersionService.
	Versions() SecretVersionService

	// Write encrypts and writes any secret data to SecretHub, always creating
	// a new secret version for the written data. This ensures secret data is
	// never overwritten.
	//
	// To ensure forward secrecy, a new secret key is used whenever the previously
	// used key has been flagged.
	//
	// Write accepts any non-empty byte data that is within the size limit of MaxSecretSize.
	// Note that data is encrypted as is. Sanitizing data is the responsibility of the
	// function caller.
	Write(path string, data []byte) (*api.SecretVersion, error)
}

func newSecretService(client client) SecretService {
	return secretService{
		client: client,
	}
}

type secretService struct {
	client client
}

// Delete removes the secret at the given path.
func (s secretService) Delete(path string) error {
	secretPath, err := api.NewSecretPath(path)
	if err != nil {
		return errio.Error(err)
	}

	secretBlindName, err := s.client.convertPathToBlindName(secretPath)
	if err != nil {
		return errio.Error(err)
	}

	err = s.client.httpClient.DeleteSecret(secretBlindName)
	if err != nil {
		return errio.Error(err)
	}

	return nil
}

// Exists returns whether a secret exists on the given path.
func (s secretService) Exists(path string) (bool, error) {
	secretPath, err := api.NewSecretPath(path)
	if err != nil {
		return false, errio.Error(err)
	}

	blindName, err := s.client.convertPathToBlindName(secretPath)
	if err != nil {
		return false, errio.Error(err)
	}

	_, err = s.client.httpClient.GetSecret(blindName)
	if err == api.ErrSecretNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// Get retrieves a Secret.
func (s secretService) Get(path string) (*api.Secret, error) {
	secretPath, err := api.NewSecretPath(path)
	if err != nil {
		return nil, errio.Error(err)
	}

	blindName, err := s.client.convertPathToBlindName(secretPath)
	if err != nil {
		return nil, errio.Error(err)
	}

	encSecret, err := s.client.httpClient.GetSecret(blindName)
	if err != nil {
		return nil, errio.Error(err)
	}

	accountKey, err := s.client.getAccountKey()
	if err != nil {
		return nil, errio.Error(err)
	}

	return encSecret.Decrypt(accountKey)
}

// Write encrypts and writes any secret data to SecretHub, always creating
// a new secret version for the written data. This ensures secret data is
// never overwritten.
//
// To ensure forward secrecy, a new secret key is used whenever the previously
// used key has been flagged.
//
// Write accepts any non-empty byte data that is within the size limit of MaxSecretSize.
// Note that data is encrypted as is. Sanitizing data is the responsibility of the
// function caller.
func (s secretService) Write(path string, data []byte) (*api.SecretVersion, error) {
	secretPath, err := api.NewSecretPath(path)
	if err != nil {
		return nil, errio.Error(err)
	}

	if secretPath.HasVersion() {
		return nil, ErrCannotWriteToVersion
	}

	if len(data) == 0 {
		return nil, ErrEmptySecret
	}

	if len(data) > MaxSecretSize {
		return nil, ErrSecretTooBig
	}

	key, err := s.client.getSecretKey(secretPath)
	if err == api.ErrSecretNotFound {
		return s.client.createSecret(secretPath, data)
	} else if err == api.ErrNoOKSecretKey {
		key, err = s.client.createSecretKey(secretPath)
		if err != nil {
			return nil, errio.Error(err)
		}
	} else if err != nil {
		return nil, errio.Error(err)
	}

	return s.client.createSecretVersion(secretPath, data, key)
}

// ListEvents retrieves all audit events for a given secret.
// If subjectTypes is left empty, the server's default is used.
func (s secretService) ListEvents(path string, subjectTypes api.AuditSubjectTypeList) ([]*api.Audit, error) {
	secretPath, err := api.NewSecretPath(path)
	if err != nil {
		return nil, errio.Error(err)
	}

	blindName, err := s.client.convertPathToBlindName(secretPath)
	if err != nil {
		return nil, errio.Error(err)
	}

	events, err := s.client.httpClient.AuditSecret(blindName, subjectTypes)
	if err != nil {
		return nil, errio.Error(err)
	}

	err = s.client.decryptAuditEvents(events...)
	if err != nil {
		return nil, errio.Error(err)
	}

	return events, nil
}

// Versions returns a SecretVersionService.
func (s secretService) Versions() SecretVersionService {
	return newSecretVersionService(s.client)
}

// convertsToBlindName will convert a path to a blindname.
func (c *client) convertPathToBlindName(path api.BlindNamePath) (string, error) {
	repoIndexKey, err := c.getRepoIndexKey(path.GetRepoPath())
	if err != nil {
		return "", errio.Error(err)
	}

	blindName, err := path.BlindName(repoIndexKey)
	if err != nil {
		return "", errio.Error(err)
	}
	return blindName, nil
}
