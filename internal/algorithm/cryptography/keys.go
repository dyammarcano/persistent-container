package cryptography

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"encoding/hex"
	"errors"
)

var g = NewKeys(keysLenght)

type Keys struct {
	length int
}

func NewKeys(length int) *Keys {
	return &Keys{length: length}
}

func Encrypt(message, password []byte) ([]byte, error) {
	return g.encrypt(message, password)
}

func Decrypt(cypherText []byte) ([]byte, error) {
	return g.decrypt(cypherText)
}

func (k *Keys) convertToKeys(inputKey []byte) ([]byte, []byte, error) {
	if len(inputKey) < 2 {
		return nil, nil, errors.New("input key too short")
	}

	selectedKey, err := k.getSpecifiedKey(inputKey)
	if err != nil {
		return nil, nil, errors.New("error generating key")
	}

	return generateKeyAndNonce(inputKey, selectedKey, k.length)
}

func (k *Keys) getSpecifiedKey(inputKey []byte) ([]byte, error) {
	index := int(binary.BigEndian.Uint16(inputKey)) % k.length
	key := stringKeys[index]
	return decodeKey(key, k.length)
}

func (k *Keys) wrapValues(cypherText, key []byte) []byte {
	return getResponseBytes(cypherText, key)
}

func (k *Keys) unwrapValues(cypherText []byte) ([]byte, []byte, []byte, error) {
	if len(cypherText) < 2 {
		return nil, nil, nil, errors.New("cypherText too short")
	}

	return unwrapKeyNonceAndMessage(cypherText, k)
}

func (k *Keys) encrypt(message, password []byte) ([]byte, error) {
	return performEncryption(message, password, k)
}

func (k *Keys) decrypt(cypherText []byte) ([]byte, error) {
	return performDecryption(cypherText, k)
}

func generateKeyAndNonce(inputKey, selectedKey []byte, length int) ([]byte, []byte, error) {
	inputKeySize := len(inputKey)
	matrixKey := make([]byte, length)

	for i, value := range selectedKey {
		matrixKey[i] = value ^ inputKey[i%inputKeySize] ^ byte(i)
	}

	key := make([]byte, 32)
	copy(key, matrixKey[:len(key)])

	nonce := make([]byte, 12)
	copy(nonce, matrixKey[len(key):len(key)+len(nonce)])

	return key, nonce, nil
}

func decodeKey(key string, length int) (keyData []byte, err error) {
	if len(key) < 2 {
		return nil, errors.New("AmountKeys is too short")
	}

	data, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}

	keyData = make([]byte, length)
	copy(keyData, data)

	return keyData, nil
}

func getResponseBytes(cypherText, key []byte) []byte {
	var response bytes.Buffer
	response.Write(key)
	response.Write(cypherText)
	response.WriteByte(byte(len(key)))

	return response.Bytes()
}

func unwrapKeyNonceAndMessage(cypherText []byte, k *Keys) ([]byte, []byte, []byte, error) {
	keyLen := int(cypherText[len(cypherText)-1])
	key, nonce, err := k.convertToKeys(cypherText[:keyLen])
	if err != nil {
		return nil, nil, nil, err
	}
	return key, nonce, cypherText[keyLen : len(cypherText)-1], nil
}

func performEncryption(message, password []byte, k *Keys) ([]byte, error) {
	key, nonce, err := k.convertToKeys(password)
	if err != nil {
		return nil, err
	}

	cc, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(cc)
	if err != nil {
		return nil, err
	}

	sealed := gcm.Seal(nil, nonce, message, nil)
	if sealed == nil {
		return nil, errors.New("encryption error")
	}

	return k.wrapValues(sealed, password), nil
}

func performDecryption(cypherText []byte, k *Keys) ([]byte, error) {
	key, nonce, message, err := k.unwrapValues(cypherText)
	if err != nil {
		return nil, err
	}

	cc, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(cc)
	if err != nil {
		return nil, err
	}

	decrypted, err := gcm.Open(nil, nonce, message, nil)
	if err != nil {
		return nil, err
	}

	if decrypted == nil {
		return nil, errors.New("decryption error")
	}

	return decrypted, nil
}
