package auth_test

import (
	"net/http"
	"testing"

	"github.com/secrethub/secrethub-go/internals/api"

	"github.com/secrethub/secrethub-go/internals/api/uuid"
	"github.com/secrethub/secrethub-go/internals/assert"
	"github.com/secrethub/secrethub-go/internals/auth"
)

func TestVerifyMultipleMethods(t *testing.T) {
	// TODO: Check whether this test still adds value, as it no longer uses multiple methods.

	// Arrange
	key := clientKey
	fingerprint, err := key.Public().Fingerprint()
	assert.OK(t, err)
	pub, err := key.Public().Export()
	assert.OK(t, err)

	fakeCredentialGetter := fakeCredentialGetter{
		GetFunc: func(arg string) (*api.Credential, error) {
			return &api.Credential{
				AccountID:   uuid.New(),
				Fingerprint: fingerprint,
				Verifier:    pub,
			}, nil
		},
	}

	cases := map[string]struct {
		Credential auth.Credential
		Expected   string
	}{
		"success": {
			Credential: auth.NewRSACredential(key),
			Expected:   fingerprint,
		},
	}

	authenticator := auth.NewAuthenticator(
		auth.NewMethodSignature(fakeCredentialGetter),
	)

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Arrange
			req, err := http.NewRequest("POST", "https://api.secrethub.io/repos/jdoe/catpictures", nil)
			assert.OK(t, err)

			err = tc.Credential.AddAuthentication(req)
			assert.OK(t, err)

			// Act
			actual, err := authenticator.Verify(req)

			// Assert
			assert.OK(t, err)

			assert.Equal(t, actual.Fingerprint, tc.Expected)
		})
	}
}
