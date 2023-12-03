package cryptography

import (
	"crypto/rand"
	"github.com/dyammarcano/base58"
)

// GenerateKeys generates a N bytes master key
func GenerateKeys(size int) ([]byte, error) {
	masterKey := make([]byte, size)
	if _, err := rand.Read(masterKey); err != nil {
		return nil, err
	}
	return masterKey, nil
}

// AutoEncryptString decrypts a message using AES-256-GCM
func AutoEncryptString(message string) (string, error) {
	generateKey, err := GenerateKeys(12)
	if err != nil {
		return "", err
	}

	encrypted, err := Encrypt([]byte(message), generateKey)
	if err != nil {
		return "", err
	}

	return base58.StdEncoding.Encode(encrypted), err
}

// AutoEncryptBytes decrypts a message using AES-256-GCM
func AutoEncryptBytes(message []byte) ([]byte, error) {
	generateKey, err := GenerateKeys(12)
	if err != nil {
		return nil, err
	}

	encrypted, err := Encrypt(message, generateKey)
	if err != nil {
		return nil, err
	}

	return []byte(base58.StdEncoding.Encode(encrypted)), nil
}

// AutoDecryptString decrypts a message using AES-256-GCM
func AutoDecryptString(message string) (string, error) {
	decoded, err := base58.StdEncoding.Decode(message)
	if err != nil {
		return "", err
	}

	decrypted, err := Decrypt(decoded)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

// AutoDecryptBytes decrypts a message using AES-256-GCM
func AutoDecryptBytes(message []byte) ([]byte, error) {
	decoded, err := base58.StdEncoding.Decode(string(message))
	if err != nil {
		return nil, err
	}

	decrypted, err := Decrypt(decoded)
	if err != nil {
		return nil, err
	}
	return decrypted, nil
}
