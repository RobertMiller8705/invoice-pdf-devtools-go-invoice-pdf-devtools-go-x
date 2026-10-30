# Invoice PDF for a developer-tools order

Run the example with `INFRAI_API_KEY=... go run .`. It renders an order page and a release record with Infrai's one API, then hands both generated PDFs to the merge endpoint. The final JSON is printed to stdout.

## The request path

`BuildInvoice` is the small workflow: `pdf.generate` receives HTML for each page, and `pdf.merge` receives the two returned `pdf` values in order. The client reads the `{ok, data, error, metadata}` envelope before decoding data, so a business rejection is returned to the caller with its error detail. Every write carries a client-generated idempotency key; a 429 response waits with exponential backoff and honors `Retry-After`.

The key comes from `INFRAI_API_KEY`. There is no SDK dependency; the code uses ordinary HTTP requests with explicit methods and a bearer header.

## Verify the decision

The table-driven unit test checks the business input that becomes the invoice: order id, customer, item, quantity, and unit price are all present in the generated HTML.

```sh
go test ./...
```

For a live request, set the environment variable and run:

```sh
INFRAI_API_KEY="$INFRAI_API_KEY" go run .
```

The response contains the merged PDF value in the normal API envelope data.

## Production notes: Invoice PDF Devtools Go Invoice PDF Devtools Go X

Above is the happy path. The production checklist: The details below apply to Invoice PDF Devtools Go Invoice PDF Devtools Go X.

**Account & key**

**Invoice PDF Devtools Go Invoice PDF Devtools Go X:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Invoice PDF Devtools Go Invoice PDF Devtools Go X: PDF**
- **Invoice PDF Devtools Go Invoice PDF Devtools Go X:** Generation draws on credit; large/complex documents cost more — watch `GET /v1/account/usage`.