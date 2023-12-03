package encoding

import (
	"bytes"
	"dataStore/internal/algorithm/compression"
	"dataStore/internal/algorithm/cryptography"
	"encoding/gob"
	"github.com/dyammarcano/base58"
)

func Serialize(message string) (string, error) {
	comp, err := compression.CompressData([]byte(message))
	if err != nil {
		return "", err
	}

	enc, err := cryptography.AutoEncryptBytes(comp)
	if err != nil {
		return "", err
	}

	return base58.StdEncoding.EncodeToString(enc), nil
}

func SerializeStruct(v any) ([]byte, error) {
	var buffer bytes.Buffer
	if err := gob.NewEncoder(&buffer).Encode(v); err != nil {
		return nil, err
	}

	comp, err := compression.CompressData(buffer.Bytes())
	if err != nil {
		return nil, err
	}

	enc, err := cryptography.AutoEncryptBytes(comp)
	if err != nil {
		return nil, err
	}

	return []byte(base58.StdEncoding.EncodeToString(enc)), nil
}

func DeserializeStruct(message string, v any) error {
	dec, err := base58.StdEncoding.DecodeString(message)
	if err != nil {
		return err
	}

	dec, err = cryptography.AutoDecryptBytes(dec)
	if err != nil {
		return err
	}

	dec, err = compression.DecompressData(dec)
	if err != nil {
		return err
	}

	return gob.NewDecoder(bytes.NewReader(dec)).Decode(v)
}

func Deserialize(message string) (string, error) {
	dec, err := base58.StdEncoding.DecodeString(message)
	if err != nil {
		return "", err
	}

	dec, err = cryptography.AutoDecryptBytes(dec)
	if err != nil {
		return "", err
	}

	dec, err = compression.DecompressData(dec)
	if err != nil {
		return "", err
	}

	return string(dec), nil
}
