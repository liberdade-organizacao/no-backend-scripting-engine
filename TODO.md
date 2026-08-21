# Plan: Accept MsgPack alongside JSON with content-negotiation (response format mirrors request format)

## Background
- Go 1.21 project, standard library `net/http`, no framework
- `vmihailenco/msgpack/v5` added as pure-Go dependency (no CGO)
- One endpoint: `POST /actions/run`. Receives a JSON request payload with the same codec-shape payload (error / result), and returns a JSON-shaped response with the same codec-shaped payload.
- **v2 change**: request format is now mirrored for the response. Request `application/json`  → response `application/json`. Request `application/msgpack` → response `application/msgpack`. Unknown Content-Type → `415 Unsupported Media Type`.
- No `Content-Type` validation in old code; no response `Content-Type` set.

## Design decisions (confirmed upfront)
- **Accept requests** with `Content-Type: application/msgpack` OR `application/json` (case-insensitive, optional `/charset=...`).
- **Reject** every other Content-Type (including missing) with `415 Unsupported Media Type`.
- Decode every request into `map[string]interface{}` (JSON-compatible types) via the selected codec.
- **Encode the response with the same codec** used to decode the request.
- Response shape is always `map[string]interface{}{"error": nil|string, "result": <any>|nil}`.
- **New package** `common/codec/codec` with a `Codec` interface and two implementations.
- **`Controller`** exposes a `codec.Codec` that drives both `decodeBody` and `encodeResponse` for the current request.

## Investigation findings
- `encoding/json` unmarshals into `map[string]interface{}` where numbers become `float64`. `msgpack` v5's default is the same `float64`. Existing consumers in `controllers.go` use `(float64)` assertions unchanged.
- MsgPack interning via `enc.UseInternedStrings(true)` reduces encoding size when the same key appears repeatedly.
- `gopher-lua.RunLuaActionTimeout` does string work — no codec-aware path needed on the response side.
- No tests cover HTTP-level handler; need a new set of HTTP-level tests.

---

## Step 1 — Add the dependency (`/go.mod`)
- `go get github.com/vmihailenco/msgpack/v5` from `/Users/calvesdasilvajunior/Developer/betty/dev/baas/scripting`.
- Run `go mod tidy` after. Confirm `go.sum` is updated. No CGO required.

---

## Step 2 — Create the `common/codec` package
File: `common/codec/codec.go`

- Define `type Codec interface { Decode(r io.Reader, v interface{}) error; Encode(w io.Writer, v interface{}) error }`.
- Implement two codecs:
  - **`jsonCodec`**: `Decode` calls `io.ReadAll(r)` then `json.Unmarshal(data, v)`. `Encode` calls `json.Marshal(v)` then `w.Write(data)`.
  - **`msgpackCodec`**: `Decode` calls `msgpack.NewDecoder(r).Decode(v)` (default behavior — numbers are `float64`). `Encode` calls `msgpack.NewEncoder(w).UseInternedStrings(true).Encode(v)`.
- Exported constructors: `NewJSON() Codec`, `NewMsgPack() Codec`.

---

## Step 3 — Wire Content-Type dispatch into the HTTP layer
File: `controllers/controllers.go`

- Add `import "strings"` alongside existing imports.
- Replace the hard-coded `json.Unmarshal(bodyBytes, &actionInfo)` read path inside the handler with a new **`decodeBody` function**:
  ```go
  func decodeBody(r *http.Request, w http.ResponseWriter, out interface{}) (codec.Codec, error) {
      ct := r.Header.Get("Content-Type")
      if idx := strings.IndexByte(ct, ';'); idx >= 0 {
          ct = strings.TrimRight(ct[:idx], " ")
      }
      ct = strings.TrimSpace(ct)
      ct = strings.ToLower(ct)
      switch ct {
      case "application/json":
          enc := codec.NewJSON()
          if err := enc.Decode(io.NewSectionReader(r.Body, 0, math.MaxInt32), out); err != nil {
              return nil, err
          }
          return enc, nil
      case "application/msgpack":
          enc := codec.NewMsgPack()
          if err := enc.Decode(io.NewSectionReader(r.Body, 0, math.MaxInt32), out); err != nil {
              return nil, err
          }
          return enc, nil
      }
      return nil, fmt.Errorf("unsupported content-type: %s", ct)
  }
  ```
- Update the handler to use `decodeBody` and call the response-encoder with the returned codec:
  - Drop the manual `json.Unmarshal(bodyBytes, ...)` block.
  - Capture `enc codec.Codec, err error := decodeBody(r, w, &actionInfo)`.
  - On `err != nil`, if `err.Error() == "unsupported content-type: ..."` → `w.WriteHeader(415)` with `{"error":"Unsupported Media Type"}`. Else `w.WriteHeader(400)` with `{"error":"Could not decode request body"}`.
  - After the action runs, serialize the response with `enc.Encode(w, ...)`.

- Add a **`encodeResponse` helper** so every path uses the same codec-shaped output:
  ```go
  func encodeResponse(w http.ResponseWriter, enc codec.Codec, status int, payload map[string]interface{}) error {
      w.Header().Set("Content-Type", enc.ContentType())
      w.WriteHeader(status)
      return enc.Encode(w, payload)
  }
  ```
- Add a tiny method on each codec to export its content type:
  ```go
  type jsonCodec struct{}
  func (jsonCodec) ContentType() string { return "application/json" }

  type msgpackCodec struct{}
  func (msgpackCodec) ContentType() string { return "application/msgpack" }
  ```

---

## Step 4 — Wire `Content-Type` for each codec into each HTTP response path
File: `controllers/controllers.go`

Every response path (invalid method, body read failure, decode failure, permission failure, script-not-found, Lua execution failure, success) must:
- Set response `Content-Type` to the *same* value as the request Content-Type.
- Use the same codec used to decode the request for encoding the response.

Implementation:
- Use the `enc codec.Codec` returned by `decodeBody`. All `w.WriteHeader(...)` and `io.WriteString(...)` calls are replaced by `encodeResponse(w, enc, status, payload)`.
- For unknown Content-Type → `415 Unsupported Media Type`. Use `{"error":"Unsupported Media Type"}` as the payload via the JSON codec (irrelevant because it will not be consumed by a MsgPack consumer).

---

## Step 5 — Add HTTP-level tests for `HandleRunAction`
File: `controllers/controllers_test.go`

Add to `controllers` package tests. No third-party libraries needed; use `httptest.NewServer`, `encoding/json`, `io.ReadAll`, and the `common/codec` package:

- **`TestHandleRunAction_Json`** — POST with `Content-Type: application/json`. Assert response status `200`, `Content-Type: application/json` header, body is equal to a known-good JSON shape.
- **`TestHandleRunAction_MsgPack`** — POST with `Content-Type: application/msgpack`, body = `codec.NewMsgPack().Encode(...)`. Assert response status `200`, `Content-Type: application/msgpack` header, body decodes via `codec.NewMsgPack().Decode(...)` to the same payload as the JSON test. **Round-trip assertion.**
- **`TestHandleRunAction_PropertiesMsgPackSmaller`** — same as MsgPack test, additionally assert `len(msgpackBody) < len(jsonBody)` for the same action data (smoke-test that interning is active).
- **`TestHandleRunAction_UnknownContentType`** — POST with `Content-Type: application/xml`. Assert status `415`.
- **`TestHandleRunAction_NoContentType`** — POST with no Content-Type header. Assert status `415`.
- **`TestHandleRunAction_BadJson`** — POST with `Content-Type: application/json` and body `{not json}`. Assert status `400`, body contains `"Could not decode request body"`.
- **`TestHandleRunAction_BadMsgPack`** — POST with `Content-Type: application/msgpack` and body `random bytes`. Assert status `400`.
- **`TestHandleRunAction_BinaryPayload`** — POST a binary MsgPack payload with the same keys as the JSON payload. Assert status `200`, `Content-Type: application/msgpack` in the response, and body decodes to an identical payload to the JSON test.

---

## Step 6 — Update `services/services.go` register if needed
Inspect current mux registration. It uses `http.HandleFunc` (from standard library). No code change because Content-Type selection is inside the handler, not routing.
- No change here. Document: "Content-Type dispatch is inside `HandleRunAction`; `services.go` is unchanged."

---

## Step 7 — Build, lint, and smoke-test
- `go build ./... && go vet ./...` — clean.
- `go test ./...` — all existing and new tests pass.
- `make test` — equivalent.
- `go build -o main.exe .` — binary compiles.
- `./main.exe up` (or via `make run`) — start the server on `127.0.0.1:7781`.

Manual verification (all should return the matching Content-Type):
```
# JSON round-trip
curl -s -i -X POST http://localhost:7781/actions/run \
     -H "Content-Type: application/json" \
     -d '{"app_id":{"$numberInt":"1"},"user_id":{"$numberInt":"1"},"action_name":"sample","action_param":""}'

# MsgPack round-trip
# build binary payload locally via Go: msgpack.Marshal(map[string]interface{}{...})
curl -s -i -X POST http://localhost:7781/actions/run \
     -H "Content-Type: application/msgpack" \
     --data-binary <msgpack-bytes.bin>

# Rejected Content-Type
curl -s -i -X POST http://localhost:7781/actions/run \
     -H "Content-Type: application/xml" \
     -d "<xml/>"
```
- All three responses must have `Content-Type` matching the request.
- Unknown Content-Type → `415 Unsupported Media Type`.

---

## Step 8 — Commit
- `go mod tidy && git diff` to inspect.
- Commit with message: `feat: accept MsgPack payloads alongside JSON with content-negotiation`.

---

## Files touched (final v2 list)
| Action | Path |
|---|---|
| Add | `common/codec/codec.go` |
| Add | `common/codec/codec_test.go` |
| Add | `go.sum` entries |
| Modify | `controllers/controllers.go` |
| Modify | `controllers/controllers_test.go` |
| Modify | `go.mod` (new dep) |
