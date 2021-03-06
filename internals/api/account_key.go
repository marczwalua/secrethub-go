package api

import (
	"net/http"

	"github.com/secrethub/secrethub-go/internals/crypto"
)

// Errors
var (
	ErrAccountNotKeyed    = errAPI.Code("account_not_keyed").StatusError("User has not yet keyed their account", http.StatusBadRequest)
	ErrAccountKeyNotFound = errAPI.Code("account_key_not_found").StatusError("User has not yet keyed their account", http.StatusNotFound)
)

// EncryptedAccountKey represents an account key encrypted with a credential.
type EncryptedAccountKey struct {
	Account             *Account
	PublicKey           []byte
	EncryptedPrivateKey crypto.CiphertextRSAAES
	Credential          *Credential
}

// CreateAccountKeyRequest contains the fields to add an account_key encrypted for a credential.
type CreateAccountKeyRequest struct {
	EncryptedPrivateKey crypto.CiphertextRSAAES
	PublicKey           []byte
}

// Validate checks whether the request is valid.
func (req CreateAccountKeyRequest) Validate() error {
	if len(req.PublicKey) == 0 {
		return ErrInvalidPublicKey
	}
	return nil
}
