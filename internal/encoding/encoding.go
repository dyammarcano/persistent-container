package encoding

import (
	"github.com/dyammarcano/base58"
	"github.com/dyammarcano/secure_message/internal/algorithm/compression"
	"github.com/dyammarcano/secure_message/internal/algorithm/cryptography"
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
