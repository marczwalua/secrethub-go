package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/keylockerbv/secrethub-go/pkg/errio"
)

var (
	// ExternalKeyLength is the key length for external keys that are used to
	// authenticate to the API and encrypt the AccountKey.
	//
	// The size of 4096 bits has been chosen because 2048 bits is assumed to
	// last up to the year 2020. Have a look at https://keylength.org
	// Additionally, because 2048 bits and 4096 bits are the most commonly used
	// key lengths (implemented in e.g. smartcards), we expect those key lengths
	// to be more future proof in terms of usability and we expect (more) efficient
	// implementations to be developed for those key lengths over less frequently
	// used ones (e.g. 3072).
	ExternalKeyLength = 4096
)

// Errors
var (
	errCrypto = errio.Namespace("crypto")

	ErrGenerateRSAKey                  = errCrypto.Code("rsa_generate_fail").Error("could not generate RSA key")
	ErrRSADecrypt                      = errCrypto.Code("rsa_decrypt_failed").ErrorPref("could not decrypt data: %s")
	ErrRSAEncrypt                      = errCrypto.Code("rsa_encrypt_failed").ErrorPref("could not encrypt data: %s")
	ErrNoKeyInFile                     = errCrypto.Code("no_key_in_file").Error("no key found in file")
	ErrMultipleKeysInFile              = errCrypto.Code("multiple_keys_in_file").Error("multiple keys found in file")
	ErrNotPKCS1Format                  = errCrypto.Code("key_wrong_format").Error("key must be in pkcs1 format")
	ErrNoPublicKeyFoundInImport        = errCrypto.Code("no_public_key_in_import").Error("no public key found in import")
	ErrMultiplePublicKeysFoundInImport = errCrypto.Code("multiple_public_keys_in_import").Error("multiple public keys found in import")
	ErrKeyTypeNotSupported             = errCrypto.Code("key_type_not_supported").Error("key type is not supported")
	ErrEmptyPassphrase                 = errCrypto.Code("empty_passphrase").Error("passphrase is empty")

	ErrEmptyPublicKey = errors.New("public key should not be empty")
)

// RSAPublicKey wraps a RSA public key
// It exposes all crypto functionality asymmetric keys.
type RSAPublicKey struct {
	publicKey *rsa.PublicKey
}

// Encrypt encrypts the data with RSA-OAEP using the RSAKey.
func (k *RSAPublicKey) Encrypt(data []byte) (CiphertextRSA, error) {
	encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, k.publicKey, data, []byte{})
	if err != nil {
		return CiphertextRSA{}, ErrRSAEncrypt(err)
	}

	return CiphertextRSA{
		Data: encrypted,
	}, nil
}

// EncryptBytes encrypts the data with RSA-OAEP using the RSAPublicKey and
// will be deprecated. Directly use Encrypt instead.
func (k *RSAPublicKey) EncryptBytes(data []byte) ([]byte, error) {
	encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, k.publicKey, data, []byte{})
	if err != nil {
		return nil, ErrRSAEncrypt(err)
	}
	return encrypted, nil
}

// Verify will verify a message using the encoded public key and the signature.
// A valid signature is indicated by returning a nil error
// A public key must be importable by importPublicKey.
func (k *RSAPublicKey) Verify(message, signature []byte) error {
	hashedMessage := sha256.Sum256(message)

	return rsa.VerifyPKCS1v15(k.publicKey, crypto.SHA256, hashedMessage[:], signature)
}

// Verify will verify a message using the encoded public key and the signature.
// A valid signature is indicated by returning a nil error
// A public key must be importable by importPublicKey.
func Verify(encodedPublicKey, message, signature []byte) error {
	publicKey, err := ImportRSAPublicKey(encodedPublicKey)
	if err != nil {
		return errio.Error(err)
	}

	return publicKey.Verify(message, signature)
}

// ExportPublicKey exports the rsa public key in an PKIX pem encoded format.
func (k RSAPublicKey) ExportPublicKey() ([]byte, error) {
	asn1, err := x509.MarshalPKIXPublicKey(k.publicKey)
	if err != nil {
		return nil, errio.Error(err)
	}

	bytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: asn1,
	})

	return bytes, nil
}

// ImportRSAPublicKey imports a RSAPublic key from an exported public key.
func ImportRSAPublicKey(encodedPublicKey []byte) (*RSAPublicKey, error) {
	if len(encodedPublicKey) == 0 {
		return nil, ErrEmptyPublicKey
	}

	pemBlock, rest := pem.Decode(encodedPublicKey)
	if pemBlock == nil {
		return nil, ErrNoPublicKeyFoundInImport
	} else if len(rest) > 0 {
		return nil, ErrMultiplePublicKeysFoundInImport
	}

	publicKey, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	if err != nil {
		return nil, ErrNotPKCS1Format
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, ErrKeyTypeNotSupported
	}

	return &RSAPublicKey{
		publicKey: rsaPublicKey,
	}, nil
}

// RSAKey wraps a RSA key.
// It exposes all crypto functionality asymmetric keys.
type RSAKey struct {
	*RSAPublicKey
	privateKey *rsa.PrivateKey
}

// NewRSAKey is used to create a new RSAKey.
// Normally you create a RSAKey by DecodeKey
func NewRSAKey(privateKey *rsa.PrivateKey) *RSAKey {
	return &RSAKey{
		RSAPublicKey: &RSAPublicKey{
			publicKey: &privateKey.PublicKey,
		},
		privateKey: privateKey,
	}
}

// GenerateRSAKey generates a new RSAKey with the given length
func GenerateRSAKey(length int) (*RSAKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, length)
	if err != nil {
		return nil, ErrGenerateRSAKey
	}

	return NewRSAKey(privateKey), nil
}

// Sign signs a message using the RSAKey.
func (k *RSAKey) Sign(message []byte) ([]byte, error) {
	hashedMessage := sha256.Sum256(message)

	return rsa.SignPKCS1v15(rand.Reader, k.privateKey, crypto.SHA256, hashedMessage[:])
}

// Decrypt decrypts the encryptedData with RSA-OAEP using the RSAKey.
func (k *RSAKey) Decrypt(ciphertext CiphertextRSA) ([]byte, error) {
	if len(ciphertext.Data) == 0 {
		return []byte{}, nil
	}

	output, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, k.privateKey, ciphertext.Data, []byte{})
	if err != nil {
		return nil, ErrRSADecrypt(err)
	}
	return output, nil
}

// ReEncrypt re-encrypts the data for the given public key.
// The RSAKey must be able to decrypt the original data for the function to succeed.
func (k *RSAKey) ReEncrypt(pk *RSAPublicKey, encData []byte) ([]byte, error) {

	decData, err := k.DecryptBytes(encData)
	if err != nil {
		return nil, errio.Error(err)
	}

	return pk.EncryptBytes(decData)
}

// DecryptBytes decrypts the encrypted data with RSA-OAEP using the RSAKey.
// This function will be deprecated. Directly use Decrypt instead.
func (k *RSAKey) DecryptBytes(encryptedData []byte) ([]byte, error) {
	output, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, k.privateKey, encryptedData, []byte{})
	if err != nil {
		return nil, ErrRSADecrypt(err)
	}
	return output, nil
}

// GenerateServiceKey generates an key pair for the Service and returns the private key and public key.
// These keys are in an exported format.
func GenerateServiceKey() (*RSAKey, error) {
	privateKey, err := GenerateRSAKey(ExternalKeyLength)
	if err != nil {
		return nil, errio.Error(err)
	}
	return privateKey, nil
}

// Fingerprint returns the SHA256 fingerprint of the public key
func (k *RSAKey) Fingerprint() (string, error) {
	pub, err := k.ExportPublicKey()
	if err != nil {
		return "", errio.Error(err)
	}

	sum := sha256.Sum256(pub)
	return hex.EncodeToString(sum[:]), nil
}

// ExportPrivateKey exports the rsa private key in an PKIX pem encoded format.
func (k RSAKey) ExportPrivateKey() ([]byte, error) {
	privateASN1 := x509.MarshalPKCS1PrivateKey(k.privateKey)

	privateBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateASN1,
	})

	return privateBytes, nil
}

// Export exports the raw rsa private key.
func (k RSAKey) Export() []byte {
	return x509.MarshalPKCS1PrivateKey(k.privateKey)
}

// ImportRSAPrivateKey imports a rsa private key from a pem encoded format.
func ImportRSAPrivateKey(privateKey []byte) (*RSAKey, error) {
	pemBlock, rest := pem.Decode(privateKey)
	if pemBlock == nil {
		return nil, ErrNoKeyInFile

	} else if len(rest) > 0 {
		return nil, ErrMultipleKeysInFile
	}

	privateRSAKey, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, ErrNotPKCS1Format
	}

	return NewRSAKey(privateRSAKey), nil
}

// ExportPrivateKeyWithPassphrase exports the rsa private key in a
// PKIX pem encoded format, encrypted with the given passphrase.
func (k RSAKey) ExportPrivateKeyWithPassphrase(pass string) ([]byte, error) {
	if pass == "" {
		return nil, ErrEmptyPassphrase
	}

	plain := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(k.privateKey),
	}

	encrypted, err := x509.EncryptPEMBlock(rand.Reader, plain.Type, plain.Bytes, []byte(pass), x509.PEMCipherAES256)
	if err != nil {
		return nil, errio.Error(err)
	}

	return pem.EncodeToMemory(encrypted), nil
}

// CiphertextRSAAES represents data encrypted with AES-GCM, where the AES-key is encrypted with RSA-OAEP.
type CiphertextRSAAES struct {
	CiphertextAES
	CiphertextRSA
}

// MarshalJSON encodes the ciphertext in a string.
func (ct CiphertextRSAAES) MarshalJSON() ([]byte, error) {
	data := base64.StdEncoding.EncodeToString(ct.CiphertextAES.Data)

	metadata := newEncodedCiphertextMetadata(map[string]string{
		"nonce": base64.StdEncoding.EncodeToString(ct.CiphertextAES.Nonce),
		"key":   base64.StdEncoding.EncodeToString(ct.CiphertextRSA.Data),
	})

	return []byte(fmt.Sprintf("%s$%s$%s", AlgorithmRSAAES, data, metadata)), nil
}

// UnmarshalJSON decodes a string into a ciphertext.
func (ct *CiphertextRSAAES) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return nil
	}

	encoded := encodedCiphertext(b)

	algorithm, err := encoded.algorithm()
	if err != nil {
		return errio.Error(err)
	}

	if algorithm != AlgorithmRSAAES {
		return ErrWrongAlgorithm
	}

	encryptedData, err := encoded.data()
	if err != nil {
		return errio.Error(err)
	}

	metadata, err := encoded.metadata()
	if err != nil {
		return errio.Error(err)
	}

	aesNonce, err := metadata.getDecodedValue("nonce")
	if err != nil {
		return errio.Error(err)
	}

	aesKey, err := metadata.getDecodedValue("key")
	if err != nil {
		return errio.Error(err)
	}

	ct.CiphertextAES = CiphertextAES{
		Data:  encryptedData,
		Nonce: aesNonce,
	}

	ct.CiphertextRSA = CiphertextRSA{
		Data: aesKey,
	}

	return nil
}

// CiphertextRSA represents data encrypted with RSA-OAEP.
type CiphertextRSA struct {
	Data []byte
}

// EncryptRSAAES encrypts provided data with AES-GCM.
// The used AES-key is then encrypted with RSA-OAEP.
func EncryptRSAAES(data []byte, k *RSAPublicKey) (CiphertextRSAAES, error) {
	aesKey, err := GenerateAESKey()
	if err != nil {
		return CiphertextRSAAES{}, errio.Error(err)
	}

	aesData, err := aesKey.Encrypt(data)
	if err != nil {
		return CiphertextRSAAES{}, errio.Error(err)
	}

	rsaData, err := k.Encrypt(aesKey.key)
	if err != nil {
		return CiphertextRSAAES{}, errio.Error(err)
	}

	return CiphertextRSAAES{
		CiphertextAES: aesData,
		CiphertextRSA: rsaData,
	}, nil
}

// DecryptRSAAES decrypts provided data that is encrypted with AES-GCM
// and then the AES-key used for encryption encrypted with RSA-OAEP.
func DecryptRSAAES(ciphertext CiphertextRSAAES, pk RSAKey) ([]byte, error) {
	return ciphertext.decrypt(pk)
}

// decrypt decrypts the key in CiphertextRSAAES with RSA-OAEP and then decrypts the data in CiphertextRSAAES with AES-GCM.
func (ct *CiphertextRSAAES) decrypt(k RSAKey) ([]byte, error) {
	aesKeyData, err := ct.CiphertextRSA.decrypt(k)
	if err != nil {
		return nil, errio.Error(err)
	}

	aesKey := &AESKey{aesKeyData}

	return aesKey.decrypt(ct.CiphertextAES.Data, ct.CiphertextAES.Nonce)
}

// decrypt decrypts the data in CiphertextRSA with RSA-OAEP using the provided key.
func (ct *CiphertextRSA) decrypt(k RSAKey) ([]byte, error) {
	if ct.Data == nil {
		return nil, ErrInvalidCiphertext
	}

	return k.DecryptBytes(ct.Data)
}

// MarshalJSON encodes the ciphertext in a string.
func (ct CiphertextRSA) MarshalJSON() ([]byte, error) {
	encodedKey := base64.StdEncoding.EncodeToString(ct.Data)

	return []byte(fmt.Sprintf("%s$%s$", AlgorithmRSA, encodedKey)), nil
}

// UnmarshalJSON decodes a string into a ciphertext.
func (ct *CiphertextRSA) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return nil
	}

	encoded := encodedCiphertext(b)

	algorithm, err := encoded.algorithm()
	if err != nil {
		return errio.Error(err)
	}

	if algorithm != AlgorithmRSA {
		return ErrWrongAlgorithm
	}

	encryptedData, err := encoded.data()
	if err != nil {
		return errio.Error(err)
	}

	ct.Data = encryptedData

	return nil
}
