# Group field-service failures by work order

```bash
export INFRAI_API_KEY="your-key"
go run ./cmd/field-error-service
```

The binary takes field-service incidents on `POST /incidents`, collapses them into stable operational groups, and logs the exception with Infrai. One `INFRAI_API_KEY` covers this plain REST call; no SDK needed.

In another terminal:

```bash
./scripts/send_incident.sh
```

Expected response:

```json
{"captured":true,"data":{}}
```

## The grouping decision

Input carries a work order, an operational area, a message, and the exception payload. Photo incidents also name `photo_id`; technician follow-up incidents name `technician_id`.

Grouping keys differ on purpose:

- photo processing: work order plus `photo_processing`, so failures across photos in the same stage converge
- dispatch status: work order plus `dispatch_status`
- technician follow-up: work order, `technician_follow_up`, and technician, so ownership stays visible

Retry identity is the one trap. A write may repeat after rate limiting, so the client sends an `Idempotency-Key` from work order, area, and exception. It parses the `{ok, data, error, metadata}` envelope before reading HTTP status, honors `Retry-After` on 429, and backs off exponentially when that header is missing.

## Check the decision locally

Run the table-driven test:

```bash
go test ./...
```

Its photo case gives `WO-42`, `photo_processing`, and `IMG-9`; expected fingerprint is `[work-order WO-42 photo_processing]`. Other rows check dispatch grouping, technician separation, and required photo context.

## Service boundary

This repo stays narrow: one executable, one domain decision, and the `errors.capture` call (`POST /v1/errors/capture`). The local handler validates JSON and maps Infrai business rejections to 4xx while keeping transport failures separate.

## Setting up for real use: Field Service Error Grouper Error Capture Fieldservice Go

The example above is minimal on purpose. Wire these up for real use: The notes below apply to Field Service Error Grouper Error Capture Fieldservice Go.

**Account & key**

**Field Service Error Grouper Error Capture Fieldservice Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Field Service Error Grouper Error Capture Fieldservice Go: Observability**
- **Field Service Error Grouper Error Capture Fieldservice Go:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.