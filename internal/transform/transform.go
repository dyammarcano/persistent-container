package transform

import (
	"bytes"
	"encoding/gob"
	flatbuffers "github.com/google/flatbuffers/go"
)

func EncodeGob(v any) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(v)
	return buf.Bytes(), err
}

func DecodeGob(data []byte, v any) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(v)
}

func DecodeGobGeneric[T any](data []byte) (T, error) {
	var v T
	err := DecodeGob(data, &v)
	return v, err
}

func EncodeFlatBuffers(v any) ([]byte, error) {
	builder := flatbuffers.NewBuilder(0)
	return builder.FinishedBytes(), nil
}

func DecodeFlatBuffers(data []byte, v any) error {
	return nil
}
