package crypto

import (
	"crypto/rand"

	"github.com/keylockerbv/secrethub/core/errio"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// GenerateNonce generates a Nonce of a particular size.
func GenerateNonce(size int) (*[]byte, error) {
	nonce := make([]byte, size)
	if _, err := rand.Read(nonce); err != nil {
		return nil, errio.Error(err)
	}
	return &nonce, nil
}