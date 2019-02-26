package crypto

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/keylockerbv/secrethub-go/pkg/testutil"
)

func TestAESKey_Encrypt_Decrypt_Secret(t *testing.T) {
	encryptionKey, err := GenerateAESKey()
	testutil.Compare(t, err, nil)

	testData := []byte("testdata")

	ciphertext, err := encryptionKey.Encrypt(testData)
	if err != nil {
		t.Error(err)
	}

	decryptedData, err := encryptionKey.Decrypt(ciphertext)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(testData, decryptedData) {
		t.Fail()
	}
}

func TestSymmetricKey_HMAC(t *testing.T) {
	// Setup
	indexKey, err := GenerateAESKey()
	testutil.OK(t, err)
	testData := []byte("testDataString")

	// Act
	result, err := indexKey.HMAC(testData)
	testutil.OK(t, err)

	// Assert
	if bytes.Equal(result, testData) {
		t.Fail()
	}

	// Hash should not be appended.
	if len(result) > len(testData) {
		if bytes.Equal(result[:len(testData)], testData) {
			t.Fatal("Hash is appended")
		}
	}
}

func TestCiphertextAES_MarshallJSON(t *testing.T) {
	cases := map[string]struct {
		ciphertext CiphertextAES
		expected   string
	}{
		"success": {
			ciphertext: CiphertextAES{
				Data:  []byte("aes_data"),
				Nonce: []byte("nonce_data"),
			},
			expected: "AES-GCM$YWVzX2RhdGE=$nonce=bm9uY2VfZGF0YQ==",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			actual, err := tc.ciphertext.MarshalJSON()
			testutil.OK(t, err)
			expected, err := json.Marshal(tc.expected)
			testutil.OK(t, err)

			// Assert
			testutil.Compare(t, actual, expected)
		})
	}
}

func TestCiphertextRSAAES_MarshalJSON(t *testing.T) {
	cases := map[string]struct {
		ciphertext CiphertextRSAAES
		expected   string
	}{
		"success": {
			ciphertext: CiphertextRSAAES{
				CiphertextAES: CiphertextAES{
					Data:  []byte("aes_data"),
					Nonce: []byte("nonce_data"),
				},
				CiphertextRSA: CiphertextRSA{
					Data: []byte("rsa_data"),
				},
			},
			expected: "RSA-OAEP+AES-GCM$YWVzX2RhdGE=$key=cnNhX2RhdGE=,nonce=bm9uY2VfZGF0YQ==",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			actual, err := tc.ciphertext.MarshalJSON()
			testutil.OK(t, err)
			expected, err := json.Marshal(tc.expected)
			testutil.OK(t, err)

			// Assert
			testutil.Compare(t, actual, expected)
		})
	}
}
