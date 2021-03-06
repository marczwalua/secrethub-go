package auth_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/secrethub/secrethub-go/internals/api"
	"github.com/secrethub/secrethub-go/internals/errio"

	"time"

	"github.com/secrethub/secrethub-go/internals/api/uuid"
	"github.com/secrethub/secrethub-go/internals/assert"
	"github.com/secrethub/secrethub-go/internals/auth"
	"github.com/secrethub/secrethub-go/internals/crypto"
)

var (
	clientKey     crypto.RSAPrivateKey
	diffClientKey crypto.RSAPrivateKey
)

func init() {
	var err error
	clientKey, err = crypto.GenerateRSAPrivateKey(1024)
	if err != nil {
		panic(err)
	}

	diffClientKey, err = crypto.GenerateRSAPrivateKey(1024)
	if err != nil {
		panic(err)
	}
}

func TestVerify(t *testing.T) {
	fingerprint1, err := clientKey.Public().Fingerprint()
	assert.OK(t, err)

	pub1, err := clientKey.Public().Export()
	assert.OK(t, err)

	key1 := &api.Credential{
		AccountID:   uuid.New(),
		Fingerprint: fingerprint1,
		Verifier:    pub1,
	}

	cases := map[string]struct {
		Authorization string
		Date          string
		Credential    *api.Credential
		GetErr        error
		Expected      *auth.Result
		Err           error
	}{
		"empty_date": {
			Authorization: "secrethub-sig-v1 foo:bar",
			Date:          "",
			Err:           auth.ErrCannotParseDateHeader,
		},
		"invalid_date_format": {
			Authorization: "secrethub-sig-v1 foo:bar",
			Date:          time.Now().Format(time.RFC3339),
			Err:           auth.ErrCannotParseDateHeader,
		},
		"empty_authorization_header": {
			Authorization: "",
			Date:          time.Now().Format(time.RFC1123),
			Err:           auth.ErrInvalidAuthorizationHeader,
		},
		"invalid_method": {
			Authorization: "Basic username:password",
			Date:          time.Now().Format(time.RFC1123),
			Err:           auth.ErrInvalidAuthorizationHeader,
		},
		"invalid_format": {
			Authorization: "secrethub-sig-v1 no_colon_here",
			Date:          time.Now().Format(time.RFC1123),
			Err:           auth.ErrInvalidAuthorizationHeader,
		},
		"too_many_colons": {
			Authorization: "secrethub-sig-v1 foo:bar:baz:extra",
			Date:          time.Now().Format(time.RFC1123),
			Err:           auth.ErrInvalidAuthorizationHeader,
		},
		"invalid_signature_format": {
			Authorization: "secrethub-sig-v1 RSA$base64_encoded_fingerprint:signature_not_base64",
			Date:          time.Now().Format(time.RFC1123),
			Err:           auth.ErrMalformedSignature,
		},
		"fingerprint_not_found": {
			Authorization: fmt.Sprintf("secrethub-sig-v1 %s:%s", key1.Fingerprint, base64.StdEncoding.EncodeToString([]byte("some_signature"))),
			Date:          time.Now().Format(time.RFC1123),
			GetErr:        api.ErrCredentialNotFound,
			Err:           api.ErrSignatureNotVerified,
		},
		"unexpected_get_error": {
			Authorization: fmt.Sprintf("secrethub-sig-v1 %s:%s", key1.Fingerprint, base64.StdEncoding.EncodeToString([]byte("some_signature"))),
			Date:          time.Now().Format(time.RFC1123),
			GetErr:        errio.Namespace("testing").Code("get_key_failed").StatusError("cannot get account key", http.StatusInternalServerError),
			Err:           errio.Namespace("testing").Code("get_key_failed").StatusError("cannot get account key", http.StatusInternalServerError),
		},
		"invalid_signature": {
			Authorization: fmt.Sprintf("secrethub-sig-v1 %s:%s", key1.Fingerprint, base64.StdEncoding.EncodeToString([]byte("some_signature"))),
			Date:          time.Now().Format(time.RFC1123),
			Credential:    key1,
			GetErr:        nil,
			Err:           api.ErrSignatureNotVerified,
		},
		"deprecated v1": {
			Authorization: "SecretHub foo:bar:baz",
			Date:          time.Now().Format(time.RFC1123),
			Err:           auth.ErrOutdatedSignatureProtocol,
		},
		"deprecated v2": {
			Authorization: "SecretHub-Sig2 foo:bar",
			Date:          time.Now().Format(time.RFC1123),
			Err:           auth.ErrOutdatedSignatureProtocol,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Arrange
			req, err := http.NewRequest("GET", "https://api.secrethub.io/repos/jdoe/catpictures", nil)
			assert.OK(t, err)

			req.Header.Set("Authorization", tc.Authorization)
			req.Header.Set("Date", tc.Date)

			fakeCredentialGetter := fakeCredentialGetter{
				GetFunc: func(fingerprint string) (*api.Credential, error) {
					return tc.Credential, tc.GetErr
				},
			}

			authenticator := auth.NewMethodSignature(fakeCredentialGetter)

			// Act
			actual, err := authenticator.Verify(req)

			// Assert
			assert.Equal(t, err, tc.Err)

			if tc.Err == nil {
				assert.Equal(t, actual, tc.Expected)
			}
		})
	}
}

func TestSignRequest(t *testing.T) {

	// Arrange
	key1 := clientKey
	fingerprint1, err := key1.Public().Fingerprint()
	assert.OK(t, err)
	pub1, err := key1.Public().Export()
	assert.OK(t, err)

	key2 := diffClientKey
	pub2, err := key2.Public().Export()
	assert.OK(t, err)

	cases := map[string]struct {
		ClientKey           crypto.RSAPrivateKey
		StoredPub           []byte
		ExpectedFingerprint string
		Err                 error
	}{
		"valid": {
			ClientKey:           key1,
			StoredPub:           pub1,
			ExpectedFingerprint: fingerprint1,
			Err:                 nil,
		},
		"pub_does_not_match_client_key": {
			ClientKey: key1,
			StoredPub: pub2,
			Err:       api.ErrSignatureNotVerified,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Arrange
			req, err := http.NewRequest("POST", "https://api.secrethub.io/repos/jdoe/catpictures", nil)
			assert.OK(t, err)

			fingerprint, err := tc.ClientKey.Public().Fingerprint()
			assert.OK(t, err)

			fakeCredentialGetter := fakeCredentialGetter{
				GetFunc: func(arg string) (*api.Credential, error) {
					return &api.Credential{
						AccountID:   uuid.New(),
						Fingerprint: fingerprint,
						Verifier:    tc.StoredPub,
					}, nil
				},
			}

			authenticator := auth.NewMethodSignature(fakeCredentialGetter)

			err = auth.NewRSACredential(tc.ClientKey).AddAuthentication(req)
			assert.OK(t, err)

			// Act
			actual, err := authenticator.Verify(req)

			// Assert
			assert.Equal(t, err, tc.Err)
			if err == nil {
				assert.Equal(t, actual.Fingerprint, tc.ExpectedFingerprint)
			}
		})
	}
}

func TestSignRequest_CheckHeadersAreSet(t *testing.T) {

	// Arrange
	req, err := http.NewRequest("GET", "https://api.secrethub.io/repos/jdoe/catpictures", nil)
	assert.OK(t, err)

	// Act
	err = auth.NewRSACredential(clientKey).AddAuthentication(req)
	assert.OK(t, err)

	// Assert
	if req.Header.Get("Date") == "" {
		t.Error("Date header not set.")
	}

	if req.Header.Get("Authorization") == "" {
		t.Error("Authorization header not set.")
	}
}

func TestReplayRequest(t *testing.T) {

	// Arrange
	fingerprint, err := clientKey.Public().Fingerprint()
	assert.OK(t, err)
	pub, err := clientKey.Public().Export()
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
	authenticator := auth.NewMethodSignature(fakeCredentialGetter)

	cases := map[string]struct {
		originalMethod string
		originalURL    string
		originalBody   io.Reader
		replayMethod   string
		replayURL      string
		replayBody     io.Reader
	}{
		"diff_route": {
			originalMethod: "GET",
			originalURL:    "https://api.secrethub.io/repos/jdoe/catpictures",
			originalBody:   nil,
			replayMethod:   "GET",
			replayURL:      "https://api.secrethub.io/repos/jdoe/different",
			replayBody:     nil,
		},
		"diff_method": {
			originalMethod: "GET",
			originalURL:    "https://api.secrethub.io/repos/jdoe/catpictures",
			originalBody:   nil,
			replayMethod:   "POST",
			replayURL:      "https://api.secrethub.io/repos/jdoe/catpictures",
			replayBody:     nil,
		},
		"diff_body": {
			originalMethod: "GET",
			originalURL:    "https://api.secrethub.io/repos/jdoe/catpictures",
			originalBody:   bytes.NewBufferString("some request body"),
			replayMethod:   "GET",
			replayURL:      "https://api.secrethub.io/repos/jdoe/catpictures",
			replayBody:     bytes.NewBufferString("different request body"),
		},
		"diff_body_empty": {
			originalMethod: "GET",
			originalURL:    "https://api.secrethub.io/repos/jdoe/catpictures",
			originalBody:   nil,
			replayMethod:   "GET",
			replayURL:      "https://api.secrethub.io/repos/jdoe/catpictures",
			replayBody:     bytes.NewBufferString("different request body"),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Arrange
			original, err := http.NewRequest(tc.originalMethod, tc.originalURL, tc.originalBody)
			assert.OK(t, err)

			err = auth.NewRSACredential(clientKey).AddAuthentication(original)
			assert.OK(t, err)

			replay, err := http.NewRequest(tc.replayMethod, tc.replayURL, tc.replayBody)
			assert.OK(t, err)

			// Copy the signed headers of the original request to the replay request.
			replay.Header = original.Header

			// Act
			_, err = authenticator.Verify(replay)

			// Assert
			assert.Equal(t, err, api.ErrSignatureNotVerified)
		})
	}
}

// Make sure new users cannot be created with a colon (:) in their username.
// Allowing colons would break the Authorization header format.
func TestNewUser_InvalidName(t *testing.T) {

	// Arrange
	invalidName := "John:Doe"

	// Act
	err := api.ValidateUsername(invalidName)

	// Assert
	assert.Equal(t, err, api.ErrInvalidUsername)
}
