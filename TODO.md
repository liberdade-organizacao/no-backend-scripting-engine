# Plan: Add MsgPack support alongside JSON

## Background

- Go 1.21 project, standard library `net/http` only (no external web framework).
- Single dependency set: `branca/v2`, `golang-jwt/v5`, `yuin/gopher-lua`, `modernc.org/sqlite`.
- One API endpoint accepts payloads: `POST /actions/run` — currently hard-coded to parse JSON via `encoding/json.Unmarshal` into `map[string]interface{}`.
- No `Content-Type` validation on request, no `Content-Type` header set on response, no `DisallowUnknownFields()`, no field-level validation (raw interface assertions → panic risk).
- Responses are raw `io.WriteString(w, ...)` with hand-assembled JSON, zero Content-Type header, zero status codes.
- Tests (3 files, 165 lines) cover controller methods directly — **no HTTP-level handler tests**.

## Design decisions (confirmed upfront)

- **Accept**: request `Content-Type: application/msgpack` OR `application/json` (case-insensitive, with optional `; charset=utf-8`).
- **Reject** everything else with `415 Unsupported Media Type`.
- **Serialize payload into** the same `map[string]interface{}` used today — no second parse path. A `Codec` interface abstracts the format.
- **Set response `Content-Type` to `application/json`** always (the action response is already JSON-shaped).
- **Add dependency**: `github.com/vmihailenco/msgpack/v5` (v5 for Go 1.18+ compat; v4 also works with 1.21 — picking v5 since it is the current maintained line and is actively used in the ecosystem).
- **New package** `common/codec` — not inside `controllers`, so HTTP concerns stay separated from codec selection.

## Investigation findings embedded in plan

- `encoding/json` uses `map[string]interface{}` when unmarshalling into `interface{}` — `application/msgpack` + `vmihailenco/msgpack/v5` produces the same result using `Decoder.Decode(&out)` with an `interface{}` target.
- Number handling: `json.Unmarshal` turns numbers into `float64` (matches current `interface{}` → `(float64)` assertions). `msgpack` v5 defaults to `float64` for numbers too, so no change needed in existing consumers.
- `gopher-lua` calls use raw `json.Unmarshal(body, &actionInfo)` then interface assertions — replacing with a codec-agnostic parse preserves compatibility if the map key/value types stay identical.
- No tests hit the HTTP handler — we need a thin HTTP-level test so regressions in decoding/encoding are caught.

---

## Steps

### Step 1 — Add the MsgPack dependency

- `go get github.com/vmihailenco/msgpack/v5` in project root.
- Confirm `go.sum` is updated; no `CGO` required (pure-Go codec).

### Step 2 — Create the codec interface (`common/codec/codec.go`)

- Define `type Codec interface { Decode(r io.Reader, v interface{}) error; Encode(w io.Writer, v interface{}) error }`.
- Create `var Default = NewJSON()` exported default (alias) so existing call sites can migrate by swapping `Default` for `msgpack.Default` downstream if ever needed.
- Two implementations:
  - `jsonCodec` — wraps `encoding/json`. `Decode` reads all of `r` then calls `json.Unmarshal(b, v)`. `Encode` calls `json.Marshal(v)`, then `w.Write(b)`.
  - `msgpackCodec` — wraps `vmihailenco/msgpack/v5`. `Decode` calls `msgpack.NewDecoder(r).Decode(v)`. `Encode` calls `msgpack.NewEncoder(w).Encode(v)`.
- New exported constructors: `NewJSON() Codec`, `NewMsgPack() Codec`.
- `NewMsgPack()` must also call `msgpack.EnableInterning()` on the decoder (and optionally encoder) so repeated string keys — the very reason to use MsgPack — are de-duplicated.
- Add unit tests in `common/codec/codec_test.go`:
  - round-trip a `map[string]interface{}` through JSON and MsgPack, confirm identical `reflect.DeepEqual` output (key order preserved).
  - confirm MsgPack binary output is smaller than JSON for the same payload (assert `len(msgpackBytes) < len(jsonBytes)`).
  - confirm `NewMsgPack()` enables interned output (encode same struct twice, confirm second encoding is shorter than non-interned equivalent — exercise by checking that the second encoding skips the key bytes).
  - confirm invalid JSON produces `*json.UnmarshalTypeError`-shaped errors and invalid MsgPack produces `*msgpack.InvalidUnmarshalError`-shaped (just check `err != nil`, non-specific — keep the test surface small).

### Step 3 — Wire Content-Type into the HTTP layer (`controllers/controllers.go`)

- In `HandleRunAction`:
  1. **Drop the existing raw `json.Unmarshal(bodyBytes, ...)` path.**
  2. Switch the Content-Type switch to a select on `r.Header.Get("Content-Type")` — strip any `; charset=...` suffix by splitting on `;` and using the first segment, lower-cased.
  3. Cases:
     - `"application/json"` → `codec.NewJSON()`.
     - `"application/msgpack"` → `codec.NewMsgPack()`.
     - anything else → set `w.WriteHeader(415)`, write `{"error":"Unsupported Media Type"}`, `return`.
  4. If the request body cannot be read, set `w.WriteHeader(400)`, write `{"error":"Could not read request body: <err>","result":null}`, return.
  5. Decode into the existing `map[string]interface{}` using `codec.Decode(r, &actionInfo)`. On error, set `w.WriteHeader(400)`, write existing `{"error":"Failed to parse JSON","result":null}` (rename to `"Could not decode request body"` below for accuracy, but keep the JSON wire shape identical).
  6. Extract `app_id`, `user_id`, `action_name`, `action_param` the same way (same interface assertions).
- **Rename the 400 response text** `"Failed to parse JSON"` to `"Could not decode request body"` — it is the honest description regardless of format. This is a wire-format change, but a rename, not a structural one.
- Add a second method `func decodeBody(r *http.Request, w http.ResponseWriter, out interface{}) (codec Codec, err error)` so the Content-Type dispatch and body-reading logic is isolated and unit-friendly — returns the selected codec so callers can inspect if needed in the future (returning `codec` is purely so the dispatch logic is self-contained and testable without stubbing a global).

### Step 4 — Set response Content-Type on all paths in `HandleRunAction`

- Before every `w.WriteHeader`/`io.WriteString`, set `w.Header().Set("Content-Type", "application/json; charset=utf-8")`.
- Apply to:
  - Invalid method (`405 Method Not Allowed`).
  - Body read failure (`400`).
  - Decode failure (`400`).
  - Permission failure (`{error, result}` JSON — currently `200`).
  - Script not found (`{error, result}` — currently `200`).
  - Lua execution failure (`{"error":"Could not run Lua script","result":null}` — currently `200`).
  - Success (`{"error":null,"result":"..."}` — currently `200`).
- This is required so clients can reliably negotiate the response format too.

### Step 5 — Add HTTP-level tests for `HandleRunAction`

- In `controllers/controllers_test.go` (reuse `setupBasicTest` helper):
  - `TestHandleRunAction_Json` — `httptest.NewServer(HandleRunAction)`, POST with `Content-Type: application/json`, assert `200`, parse response JSON, confirm fields.
  - `TestHandleRunAction_MsgPack` — same but `Content-Type: application/msgpack`, using `msgpack.Marshal` to build the body. Assert `200` and identical result to JSON test.
  - `TestHandleRunAction_PropertiesMsgPackSmaller` — assert `len(msgpackBody) < len(jsonBody)` for the same payload (smoke test that the codec is actually wired up).
  - `TestHandleRunAction_UnknownContentType` — send `Content-Type: application/xml`, assert `415`.
  - `TestHandleRunAction_NoContentType` — send body with no Content-Type header, assert `415`.
  - `TestHandleRunAction_BadJson` — send `{not json}`, assert `400` with decode-error message.
  - `TestHandleRunAction_BadMsgPack` — send random bytes with MsgPack Content-Type, assert `400`.
  - `TestHandleRunAction_BinaryPayload` — confirm binary MsgPack payload containing the same keys is accepted and produces the same result.
- Existing controller-unit tests are **not touched** — they exercise `CheckPermission` and `RunAction` directly without an HTTP layer, and remain valid.

### Step 6 — Update `services/services.go` register if needed

- Inspect current mux registration. If the mux is `http.DefaultServeMux`, no change — Content-Type selection happens inside the handler, not in routing.
- If a non-default mux is used (unlikely given current code; `services.go:22-23` uses `http.HandleFunc`), no change required. Document in step if any routing-level refactor is needed.

### Step 7 — Run tests, lint, build

- `make test` — all existing tests pass.
- `go vet ./...` — no new warnings.
- `make build` — binary compiles.
- `make docker-build` — Docker image builds.
- `make run` (or `./main.exe up`):
  - Verify existing JSON client: `curl -X POST http://localhost:7781/actions/run -H "Content-Type: application/json" -d '{"app_id":"A-1","user_id":"U-1","action_name":"sample","action_param":""}'` — should return same JSON as before.
  - Verify MsgPack client: build the request body with a small Go snippet or with `msgpack-cli`, send with `Content-Type: application/msgpack`, confirm identical response body (parse both with `jq` or equivalent).
  - Verify rejected Content-Type: send `application/xml`, confirm `415`.

### Step 8 — Inspect and commit

- Inspect `git diff` to confirm: new file `common/codec/codec.go`, new file `common/codec/codec_test.go`, modified `controllers/controllers.go` (with Content-Type dispatch and response Content-Type), modified `controllers/controllers_test.go` (HTTP-level tests), updated `go.mod`/`go.sum`.
- Commit with message `feat: accept MsgPack payloads alongside JSON`.

---

## Files to touch (estimated)

| Action | Path |
|---|---|
| Add | `common/codec/codec.go` |
| Add | `common/codec/codec_test.go` |
| Add | `go.sum` entries (via dependency install) |
| Modify | `controllers/controllers.go` |
| Modify | `controllers/controllers_test.go` |
| Modify | `go.mod` (new dep) |
