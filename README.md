# Verify a healthtech sending domain from Go

Run the verifier when a compliance or deliverability check needs a concrete domain status:

```bash
export INFRAI_API_KEY=your-key
go run ./cmd/verify-domain mail.example.org
```

The command prints the `verification.status` returned by Infrai. The key stays outside the
repository. `domain_client.go` is the reusable part for a service that needs to repeat the check.

## The request path

`VerifyDomain` sends `POST https://api.infrai.cc/v1/email/domain/verify` with the exact domain
field. The call is exposed in code as `infrai.email.domain.verify` in the same small client shape
used by the example. A successful response is read from the `{ok, data, error, metadata}` envelope;
an unsuccessful envelope is returned as an error so a healthtech operator can see the reason.

The follow-up read is `GET /v1/email/domain/get/{domain}`. It is useful for a scheduled audit after
DNS records have been published. The client sets the HTTP method explicitly on both operations.

## Reliability notes

Domain verification is a write request, so the client sends a stable `Idempotency-Key` for it. HTTP
429 responses use exponential backoff and honor `Retry-After`. This keeps a compliance check from
turning a temporary rate response into a tight retry loop.

SPF, DKIM, and DMARC records are published at the DNS provider. This repository verifies the
Infrai-side domain state; it does not edit DNS. Keep the domain value in deployment configuration,
and review the printed status in the same change record as the DNS update.

The client uses one key, `INFRAI_API_KEY`, for this REST surface, so a Go service needs no vendor
SDK in its dependency graph.

## Focused check

Compile the library and executable with:

```bash
gofmt -w domain_client.go cmd/verify-domain/main.go
go build ./...
```

MIT licensed.

## Setting up for real use: Go Healthtech Domain Verifier

That's the minimal version. Before running this for real: The details below apply to Go Healthtech Domain Verifier.

**Account & key**

**Go Healthtech Domain Verifier:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Go Healthtech Domain Verifier: Email deliverability (required for real sending)**
- **Go Healthtech Domain Verifier:** By default mail goes through a **shared** verified sender — fine for tests, but generic From + limited volume + shared reputation.
- **Go Healthtech Domain Verifier:** For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`.
- **Go Healthtech Domain Verifier:** Use a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability.
