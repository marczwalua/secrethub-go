package randchar

import (
	"crypto/rand"
	"math/big"

	"github.com/secrethub/secrethub-go/internals/errio"
)

var (
	// charsetAlphanumeric is the default pattern of characters used to generate random secrets.
	charsetAlphanumeric = []byte(`0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`)
	// charsetSymbols is added to the pattern when generator.useSymbols is true.
	charsetSymbols = []byte(`!@#$%^*-_+=.,?`)
)

// Generator generates random byte arrays.
type Generator interface {
	Generate(length int) ([]byte, error)
}

// NewGenerator creates a new random generator.
func NewGenerator(useSymbols bool) Generator {
	return &generator{
		useSymbols: useSymbols,
	}
}

type generator struct {
	useSymbols bool
}

// Generate returns a random byte array of given length.
func (generator generator) Generate(length int) ([]byte, error) {
	charset := charsetAlphanumeric
	if generator.useSymbols {
		charset = append(charset, charsetSymbols...)
	}
	return randFromCharset(charset, length)
}

func randFromCharset(charset []byte, length int) ([]byte, error) {
	data := make([]byte, length)

	lengthCharset := len(charset)
	for i := 0; i < length; i++ {
		c, err := rand.Int(rand.Reader, big.NewInt(int64(lengthCharset)))
		if err != nil {
			return nil, errio.Error(err)
		}
		data[i] = charset[c.Int64()]
	}
	return data, nil
}
