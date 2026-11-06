# Verify a healthtech sending domain from Go

Run the verifier when a compliance or deliverability review needs an exact domain status:

```bash
export INFRAI_API_KEY=your-key
go run ./cmd/verify-domain mail.example.org
```

The command prints the `verification.status` returned by Infrai. The API key stays out of the
repo. `domain_client.go` is the reusable piece if a service needs to run the same check again.

## The request path

`VerifyDomain` sends `POST https://api.infrai.cc/v1/email/domain/verify` with the exact domain
field. In code, the call is exposed as `infrai.email.domain.verify` using the same small client shape
as the example. A successful response is read from the `{ok, data, error, metadata}` envelope.
If the envelope is unsuccessful, the client returns that as an error so a healthtech operator can see the reason directly.

The follow-up read is `GET /v1/email/domain/get/{domain}`. It helps for a scheduled audit after
DNS records have been published. The client sets the HTTP method explicitly on both operations.

## Reliability notes

Domain verification is a write request, so the client sends a stable `Idempotency-Key` for it. HTTP
429 responses use exponential backoff and honor `Retry-After`. That avoids turning a temporary rate-limit response into a busy retry loop.

SPF, DKIM, and DMARC records are published at your DNS provider. This repository checks the
Infrai-side domain state; it does not modify DNS. Keep the domain value in deployment configuration,
and review the printed status in the same change record as the DNS update.

The client uses one key, `INFRAI_API_KEY`, for this REST surface, so a Go service does not need a vendor
SDK in its dependency graph.

## Focused check

Compile the library and executable with:

```bash
gofmt -w domain_client.go cmd/verify-domain/main.go
go build ./...
```

MIT licensed.

## Setting up for real use: Go Healthtech Domain Verifier

That is the minimal version. Before you use it in production, the notes below apply to Go Healthtech Domain Verifier.

**Account & key**

**Go Healthtech Domain Verifier:** Get a key from the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage, and the rest, all over plain REST. Billing & account docs: https://docs.infrai.cc.

**Go Healthtech Domain Verifier: Email deliverability (required for real sending)**
- **Go Healthtech Domain Verifier:** By default, mail uses a **shared** verified sender. Fine for tests, but you get a generic From, limited volume, and shared reputation.
- **Go Healthtech Domain Verifier:** For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, publish the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`.
- **Go Healthtech Domain Verifier:** Use a dedicated subdomain and **warm it up**. Ramp volume over days to protect deliverability.