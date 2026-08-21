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
	// ContentType returns the MIME type for this codec (e.g., "application/json").
	ContentType() string
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

func (c jsonCodec) Decode(r io.Reader, v interface{}) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (c jsonCodec) Encode(w io.Writer, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func (c jsonCodec) ContentType() string {
	return "application/json"
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

func (c msgpackCodec) ContentType() string {
	return "application/msgpack"
}
