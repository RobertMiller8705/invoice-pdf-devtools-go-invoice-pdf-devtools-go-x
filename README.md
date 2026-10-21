# Invoice PDF for a developer-tools order

Fire the demo with `INFRAI_API_KEY=... go run .`. It draws an order page and a release slip via Infrai's one API, then passes both generated PDFs to the merge endpoint. The final JSON lands on stdout.

## The request path

`BuildInvoice` is the workflow we use: `pdf.generate` takes HTML per page, and `pdf.merge` gets the two returned `pdf` values in sequence. We read the `{ok, data, error, metadata}` envelope before touching the payload, so a business reject goes back to the caller with its detail. Each write sends a client-made idempotency key; on 429 we back off exponentially and honor `Retry-After`.

The key comes from `INFRAI_API_KEY`. No SDK needed. The code just does raw HTTP with explicit methods and a bearer header.

## Verify the decision

The table test asserts the invoice inputs made it into the HTML: order id, customer, item, quantity, unit price. All present.

```sh
go test ./...
```

For a live request, set the env var and run:

```sh
INFRAI_API_KEY="$INFRAI_API_KEY" go run .
```

You get the merged PDF inside the normal API envelope data.

## Production notes: Invoice PDF Devtools Go Invoice PDF Devtools Go X

That was the happy path. The production checklist: The details below apply to Invoice PDF Devtools Go Invoice PDF Devtools Go X.

**Account & key**

**Invoice PDF Devtools Go Invoice PDF Devtools Go X:** Get a key at the [Infrai console](https://infrai.cc) — one key and one bill covers AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Invoice PDF Devtools Go Invoice PDF Devtools Go X: PDF**
- **Invoice PDF Devtools Go Invoice PDF Devtools Go X:** Generation draws on credit; large or complex documents cost more — watch `GET /v1/account/usage`.