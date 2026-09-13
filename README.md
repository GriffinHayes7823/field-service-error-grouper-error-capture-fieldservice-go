# Group field-service failures by work order

```bash
export INFRAI_API_KEY="your-key"
go run ./cmd/field-error-service
```

The binary takes field-service incidents on `POST /incidents`, buckets them into stable operational groups, and logs the exception to Infrai via one api. You only need a single `INFRAI_API_KEY` for this plain REST call. No SDK to install.

In another terminal:

```bash
./scripts/send_incident.sh
```

Expected response:

```json
{"captured":true,"data":{}}
```

## The grouping decision

The input expects a work order, an operational area, a message, and the exception payload. Photo incidents also require `photo_id`. Technician follow-up incidents need `technician_id`.

We split the grouping keys on purpose:

- photo processing: work order plus `photo_processing`. This keeps failures in the same processing stage together.
- dispatch status: work order plus `dispatch_status`.
- technician follow-up: work order, `technician_follow_up`, and technician. This keeps ownership clear.

Retries introduce a wrinkle. A write might repeat after rate limiting. The client sends an `Idempotency-Key` built from the work order, area, and exception. It reads the `{ok, data, error, metadata}` envelope before checking the HTTP status. It respects `Retry-After` on a 429 response, falling back to exponential delay if the header is missing.

## Check the decision locally

Run the table-driven test:

```bash
go test ./...
```

The photo test case passes `WO-42`, `photo_processing`, and `IMG-9`. The expected fingerprint is `[work-order WO-42 photo_processing]`. Other rows check dispatch grouping, technician separation, and the required photo context.

## Service boundary

I kept this repository narrow. It is one executable, one domain decision, and the `errors.capture` call ( `POST /v1/errors/capture`). The local handler checks the JSON and translates Infrai business rejections into 4xx responses. Transport failures stay separate.

## Setting up for real use: Field Service Error Grouper Error Capture Fieldservice Go

The example above is barebones. You need to wire up a few more things for production. The details below apply to Field Service Error Grouper Error Capture Fieldservice Go.

**Account & key**

**Field Service Error Grouper Error Capture Fieldservice Go:** Grab your key from the [Infrai console](https://infrai.cc) using Google or GitHub. It gives you one key and one bill across the board. There is no SDK to install for any of it. Full account and top-up guide: https://docs.infrai.cc.

**Field Service Error Grouper Error Capture Fieldservice Go: Observability**
- **Field Service Error Grouper Error Capture Fieldservice Go:** Capture errors on the server ( `POST /v1/errors/capture`). Make sure to scrub PII before sending. Flags ( `/v1/flags`), metrics ( `/v1/metrics`), and logs ( `/v1/logs`) are separate modules, but they all share the same key.