package encoding

import (
	"bytes"
	"encoding/gob"
	"github.com/dyammarcano/base58"
	"github.com/dyammarcano/persistent-container/internal/algorithm/compression"
	"github.com/dyammarcano/persistent-container/internal/algorithm/cryptography"
	"github.com/fxamacker/cbor/v2"
)

func Serialize(message []byte) (string, error) {
	comp, err := compression.CompressData(message)
	if err != nil {
		return "", err
	}

	enc, err := cryptography.AutoEncryptBytes(comp)
	if err != nil {
		return "", err
	}

	return base58.StdEncoding.EncodeToString(enc), nil
}

func SerializeStructGob(v any) (string, error) {
	var buffer bytes.Buffer
	if err := gob.NewEncoder(&buffer).Encode(v); err != nil {
		return "", err
	}

	return Serialize(buffer.Bytes())
}

func SerializeStructCBOR(v any) (string, error) {
	data, err := cbor.Marshal(v)
	if err != nil {
		return "", err
	}

	return Serialize(data)
}

func DeserializeStructGob(message string, v any) error {
	dec, err := Deserialize(message)
	if err != nil {
		return err
	}

	return gob.NewDecoder(bytes.NewReader(dec)).Decode(v)
}

func DeserializeStructCBOR(message string, v any) error {
	data, err := Deserialize(message)
	if err != nil {
		return err
	}

	if err = cbor.Unmarshal(data, v); err != nil {
		return err
	}

	return nil
}

func Deserialize(message string) ([]byte, error) {
	dec, err := base58.StdEncoding.DecodeString(message)
	if err != nil {
		return nil, err
	}

	dec, err = cryptography.AutoDecryptBytes(dec)
	if err != nil {
		return nil, err
	}

	dec, err = compression.DecompressData(dec)
	if err != nil {
		return nil, err
	}

	return dec, nil
}
