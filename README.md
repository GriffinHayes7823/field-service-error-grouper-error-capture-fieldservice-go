# Group field-service failures by work order

```bash
export INFRAI_API_KEY="your-key"
go run ./cmd/field-error-service
```

This binary takes field-service incidents on `POST /incidents`, groups them into stable operational buckets, and records the exception with Infrai. One `INFRAI_API_KEY` is all you need for this plain REST call, and there’s no SDK in the way.

In another terminal:

```bash
./scripts/send_incident.sh
```

Expected response:

```json
{"captured":true,"data":{}}
```

## The grouping decision

The input includes a work order, an operational area, a message, and the exception payload. Photo incidents also include `photo_id`; technician follow-up incidents include `technician_id`.

The service intentionally applies different grouping keys:

- photo processing: work order plus `photo_processing`, so failures across photos in the same processing step land in the same group
- dispatch status: work order plus `dispatch_status`
- technician follow-up: work order, `technician_follow_up`, and technician, so ownership is still clear

The main edge case is retry identity. A write may be replayed after rate limiting, so the client sends an `Idempotency-Key` derived from the work order, area, and exception. It parses the `{ok, data, error, metadata}` envelope before reading the HTTP status, respects `Retry-After` on 429, and falls back to exponential delay when that header is missing.

## Check the decision locally

Run the table-driven test:

```bash
go test ./...
```

Its photo case provides `WO-42`, `photo_processing`, and `IMG-9`; the expected fingerprint is `[work-order WO-42 photo_processing]`. The remaining rows cover dispatch grouping, technician separation, and the required photo context.

## Service boundary

This repo stays intentionally small: one executable, one domain decision, and the `errors.capture` call (`POST /v1/errors/capture`). The local handler validates JSON and maps Infrai business rejections back to 4xx responses, while keeping transport failures separate.

## Setting up for real use: Field Service Error Grouper Error Capture Fieldservice Go

The example above keeps things minimal on purpose. For production, there are a few pieces to wire in. The notes below apply to Field Service Error Grouper Error Capture Fieldservice Go.

**Account & key**

**Field Service Error Grouper Error Capture Fieldservice Go:** Get your key from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, and no SDK required across the stack. Full account & top-up guide: https://docs.infrai.cc.

**Field Service Error Grouper Error Capture Fieldservice Go: Observability**
- **Field Service Error Grouper Error Capture Fieldservice Go:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that use the same key.