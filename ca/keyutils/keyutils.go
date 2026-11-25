// Package keyutils serves as a utility to parse, encrypt and decrypt
// PKCS#1 and PKCS#8 private keys based on current FIPS mode status,
// supporting only EC type keys. It always allows PKCS#8 private keys
// and disallow PKCS#1 private keys in FIPS-mode.
package keyutils

import (
	"crypto"
	"crypto/aes"
	"crypto/ecdsa"
	cryptorand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"

	"github.com/cloudflare/cfssl/helpers"
	"github.com/emmansun/gmsm/pkcs"
)

// Formatter provides an interface for converting keys to the right format, and encrypting and decrypting keys
type Formatter interface {
	ParsePrivateKeyPEMWithPassword(pemBytes, password []byte) (crypto.Signer, error)
	DecryptPEMBlock(block *pem.Block, password []byte) ([]byte, error)
	EncryptPEMBlock(data, password []byte) (*pem.Block, error)
}

// ErrFIPSUnsupportedKeyFormat is returned when encryption/decryption operations are attempted on a PKCS1 key
// when FIPS mode is enabled.
var ErrFIPSUnsupportedKeyFormat = errors.New("unsupported key format due to FIPS compliance")

// Default is the default key util, where FIPS is not required
var Default Formatter = &utils{fips: false}

// FIPS is the key utility which enforces FIPS compliance
var FIPS Formatter = &utils{fips: true}

type utils struct {
	fips bool
}

// IsPKCS8 returns true if the provided der bytes is encrypted/unencrypted PKCS#8 key
func IsPKCS8(derBytes []byte) bool {
	if _, err := x509.ParsePKCS8PrivateKey(derBytes); err == nil {
		return true
	}

	_, encrypted := encryptedPKCS8(derBytes)
	return encrypted
}

// IsEncryptedPEMBlock checks if a PKCS#1 or PKCS#8 PEM-block is encrypted or not
func IsEncryptedPEMBlock(block *pem.Block) bool {
	_, encrypted := encryptedPKCS8(block.Bytes)
	return encrypted || x509.IsEncryptedPEMBlock(block)
}

// ParsePrivateKeyPEMWithPassword parses an encrypted or a decrypted PKCS#1 or PKCS#8 PEM to crypto.Signer.
// It returns an error in FIPS mode if PKCS#1 PEM bytes are passed.
func (u *utils) ParsePrivateKeyPEMWithPassword(pemBytes, password []byte) (crypto.Signer, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("Could not parse PEM")
	}

	if IsPKCS8(block.Bytes) {
		der := block.Bytes
		if _, encrypted := encryptedPKCS8(der); encrypted {
			var err error
			der, err = decryptPKCS8(der, password)
			if err != nil {
				return nil, err
			}
		}
		key, err := x509.ParsePKCS8PrivateKey(der)
		if err != nil {
			return nil, err
		}
		signer, ok := key.(crypto.Signer)
		if !ok {
			return nil, errors.New("pkcs8: private key does not implement crypto.Signer")
		}
		return signer, nil
	} else if u.fips {
		return nil, ErrFIPSUnsupportedKeyFormat
	}

	return helpers.ParsePrivateKeyPEMWithPassword(pemBytes, password)
}

// DecryptPEMBlock requires PKCS#1 or PKCS#8 PEM Block and password to decrypt and return unencrypted der []byte
// It returns an error in FIPS mode when PKCS#1 PEM Block is passed.
func (u *utils) DecryptPEMBlock(block *pem.Block, password []byte) ([]byte, error) {
	if IsPKCS8(block.Bytes) {
		return decryptPKCS8(block.Bytes, password)
	} else if u.fips {
		return nil, ErrFIPSUnsupportedKeyFormat
	}

	return x509.DecryptPEMBlock(block, password)
}

// EncryptPEMBlock takes DER-format bytes and password to return an encrypted PKCS#1 or PKCS#8 PEM-block
// It returns an error in FIPS mode when PKCS#1 PEM bytes are passed.
func (u *utils) EncryptPEMBlock(data, password []byte) (*pem.Block, error) {
	if IsPKCS8(data) {
		encrypter := pkcs.NewPBESEncrypter(pkcs.AES256CBC, pkcs.NewPBKDF2Opts(pkcs.SHA1, 8, 2048))
		algorithm, encrypted, err := encrypter.Encrypt(cryptorand.Reader, password, data)
		if err != nil {
			return nil, err
		}
		der, err := asn1.Marshal(encryptedPrivateKeyInfo{EncryptionAlgorithm: *algorithm, EncryptedData: encrypted})
		if err != nil {
			return nil, err
		}
		return &pem.Block{Type: "ENCRYPTED PRIVATE KEY", Bytes: der}, nil
	} else if u.fips {
		return nil, ErrFIPSUnsupportedKeyFormat
	}

	cipherType := x509.PEMCipherAES256
	return x509.EncryptPEMBlock(cryptorand.Reader,
		"EC PRIVATE KEY",
		data,
		password,
		cipherType)
}

// encryptedPrivateKeyInfo represents the PKCS#8 EncryptedPrivateKeyInfo envelope.
type encryptedPrivateKeyInfo struct {
	EncryptionAlgorithm pkix.AlgorithmIdentifier
	EncryptedData       []byte
}

func encryptedPKCS8(der []byte) (encryptedPrivateKeyInfo, bool) {
	var info encryptedPrivateKeyInfo
	rest, err := asn1.Unmarshal(der, &info)
	return info, err == nil && len(rest) == 0 && len(info.EncryptionAlgorithm.Algorithm) != 0
}

// decryptPKCS8 enforces SwarmKit's existing PBES2/PBKDF2/AES-256-CBC policy.
// Key derivation, cipher operations and padding validation are delegated to pkcs.
func decryptPKCS8(der, password []byte) ([]byte, error) {
	info, ok := encryptedPKCS8(der)
	if !ok || !pkcs.IsPBES2(info.EncryptionAlgorithm) {
		return nil, errors.New("pkcs8: only PBES2 supported")
	}
	var params pkcs.PBES2Params
	rest, err := asn1.Unmarshal(info.EncryptionAlgorithm.Parameters.FullBytes, &params)
	if err != nil || len(rest) != 0 {
		return nil, errors.New("pkcs8: invalid PBES2 parameters")
	}
	if !params.KeyDerivationFunc.Algorithm.Equal(pkcs.NewPBKDF2Opts(pkcs.SHA1, 8, 2048).OID()) {
		return nil, errors.New("pkcs8: only PBKDF2 supported")
	}
	if !params.EncryptionScheme.Algorithm.Equal(pkcs.AES256CBC.OID()) {
		return nil, errors.New("pkcs8: only AES-256-CBC supported")
	}

	// Reject malformed parameters before calling the cipher implementation.
	var iv []byte
	rest, err = asn1.Unmarshal(params.EncryptionScheme.Parameters.FullBytes, &iv)
	if err != nil || len(rest) != 0 || len(iv) != aes.BlockSize || len(info.EncryptedData) == 0 || len(info.EncryptedData)%aes.BlockSize != 0 {
		return nil, errors.New("pkcs8: invalid AES-CBC parameters")
	}
	var kdf struct {
		Salt           []byte
		IterationCount int
		KeyLength      int                      `asn1:"optional"`
		PRF            pkix.AlgorithmIdentifier `asn1:"optional"`
	}
	rest, err = asn1.Unmarshal(params.KeyDerivationFunc.Parameters.FullBytes, &kdf)
	if err != nil || len(rest) != 0 || len(kdf.Salt) == 0 || kdf.IterationCount <= 0 || (kdf.KeyLength != 0 && kdf.KeyLength != 32) {
		return nil, errors.New("pkcs8: invalid PBKDF2 parameters")
	}
	// HMAC-SHA1 is the default PRF and was the only PRF used by the old implementation.
	if len(kdf.PRF.Algorithm) != 0 && !kdf.PRF.Algorithm.Equal(asn1.ObjectIdentifier{1, 2, 840, 113549, 2, 7}) {
		return nil, errors.New("pkcs8: only PBKDF2-HMAC-SHA1 supported")
	}

	key, _, err := params.Decrypt(password, info.EncryptedData)
	if err != nil {
		return nil, err
	}
	if _, err := x509.ParsePKCS8PrivateKey(key); err != nil {
		return nil, errors.New("pkcs8: incorrect password")
	}
	return key, nil
}

// ECToPKCS8PEM converts a SEC1 EC private key to an unencrypted PKCS#8 PEM.
func ECToPKCS8PEM(inPEM []byte) ([]byte, error) {
	block, _ := pem.Decode(inPEM)
	if block == nil {
		return nil, errors.New("invalid pem bytes")
	}
	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), nil
}

// PKCS8ToECPEM converts an unencrypted PKCS#8 EC private key to a SEC1 PEM.
func PKCS8ToECPEM(inPEM []byte) ([]byte, error) {
	block, _ := pem.Decode(inPEM)
	if block == nil {
		return nil, errors.New("invalid pem bytes")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ec, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("pkcs8: private key is not ECDSA")
	}
	der, err := x509.MarshalECPrivateKey(ec)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der}), nil
}
