package codec

import (
	"encoding/json"
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

// Codec defines the interface for encoding and decoding payloads.
type Codec interface {
	Decode(r io.Reader, v interface{}) error
	Encode(w io.Writer, v interface{}) error
}

// NewJSON returns a codec that uses the encoding/json standard library.
func NewJSON() Codec {
	return jsonCodec{}
}

// NewMsgPack returns a codec that uses MsgPack with string interning enabled.
func NewMsgPack() Codec {
	return msgpackCodec{}
}

type jsonCodec struct{}

func (jsonCodec) Decode(r io.Reader, v interface{}) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (jsonCodec) Encode(w io.Writer, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

type msgpackCodec struct{}

func (c msgpackCodec) Decode(r io.Reader, v interface{}) error {
	dec := msgpack.NewDecoder(r)
	return dec.Decode(v)
}

func (c msgpackCodec) Encode(w io.Writer, v interface{}) error {
	enc := msgpack.NewEncoder(w)
	enc.UseInternedStrings(true)
	return enc.Encode(v)
}
