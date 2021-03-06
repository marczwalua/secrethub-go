package auth_test

import "github.com/secrethub/secrethub-go/internals/api"

type fakeCredentialGetter struct {
	GetFunc func(fingerprint string) (*api.Credential, error)
}

func (g fakeCredentialGetter) GetCredential(fingerprint string) (*api.Credential, error) {
	return g.GetFunc(fingerprint)
}
