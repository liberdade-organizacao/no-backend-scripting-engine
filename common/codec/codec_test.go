package codec

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestCodecRoundTrip(t *testing.T) {
	payload := map[string]interface{}{
		"app_id":     float64(1),
		"user_id":    "U-1",
		"action_name": "sample",
		"action_param": map[string]interface{}{
			"key": "value",
		},
	}

	formats := []struct {
		name  string
		codec Codec
		label string
	}{
		{"JSON", NewJSON(), "json"},
		{"MsgPack", NewMsgPack(), "msgpack"},
	}

	for _, tc := range formats {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := tc.codec.Encode(&buf, payload); err != nil {
				t.Fatalf("[%s] encode failed: %v", tc.label, err)
			}

			var decoded map[string]interface{}
			if err := tc.codec.Decode(bytes.NewReader(buf.Bytes()), &decoded); err != nil {
				t.Fatalf("[%s] decode failed: %v", tc.label, err)
			}

			if !reflect.DeepEqual(decoded, payload) {
				t.Errorf("[%s] decoded payload does not match original\n  got: %s\n  want: %s",
					tc.label, mustMarshal(decoded), mustMarshal(payload))
			}
		})
	}
}

func TestMsgPackSmallerThanJSON(t *testing.T) {
	payload := map[string]interface{}{
		"app_id":     float64(1),
		"user_id":    "U-1",
		"action_name": "sample_action_with_some_length",
		"action_param": map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
			"key4": "value4",
		},
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json marshal failed: %v", err)
	}

	var msgpackBytes bytes.Buffer
	mpCodec := NewMsgPack()
	if err := mpCodec.Encode(&msgpackBytes, payload); err != nil {
		t.Fatalf("msgpack encode failed: %v", err)
	}

	if len(msgpackBytes.Bytes()) >= len(jsonBytes) {
		t.Errorf("MsgPack payload (%d bytes) should be smaller than JSON (%d bytes)",
			len(msgpackBytes.Bytes()), len(jsonBytes))
	}
}

func TestMsgPackInterning(t *testing.T) {
	payload := map[string]interface{}{}
	for i := 0; i < 5; i++ {
		key := "repeated_key"
		payload[key] = float64(i)
	}

	var buf bytes.Buffer
	mpCodec := NewMsgPack()
	if err := mpCodec.Encode(&buf, payload); err != nil {
		t.Fatalf("msgpack encode failed: %v", err)
	}
	interned := buf.Bytes()

	if len(interned) == 0 {
		t.Fatal("interned encoding is empty")
	}
}

func TestJSONDecodeError(t *testing.T) {
	c := NewJSON()
	var v map[string]interface{}
	err := c.Decode(bytes.NewReader([]byte("{not json}")), &v)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestMsgPackDecodeError(t *testing.T) {
	c := NewMsgPack()
	var v map[string]interface{}
	err := c.Decode(bytes.NewReader([]byte("random garbage bytes!")), &v)
	if err == nil {
		t.Error("expected error for invalid MsgPack")
	}
}

func mustMarshal(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
