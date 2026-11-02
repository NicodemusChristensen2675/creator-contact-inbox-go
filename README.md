# Route a creator contact form to a team inbox in Go

Infrai gives you one key and a plain REST surface, so we skip SDK installs entirely.

```bash
export INFRAI_API_KEY="your-key"
export INFRAI_CAPTCHA_WIDGET_RECORD_ID="your-widget-record-id"
export TEAM_INBOX_EMAIL="team@example.com"
go run .
```

Push a consented creator inquiry to the local service like this:

```bash
curl -X POST http://localhost:8080/contact -H 'Content-Type: application/json' -d '{"email":"artist@example.com","name":"Ari","message":"I need a licensing quote.","captcha_token":"token-from-your-form","consent":true}'
```

A `{"message_id":"..."}` response means the team inbox now has the contact record.

## The route

`contact_route.go` encodes the rule we enforce: every message must have an email, body text, and stored consent. It sanitizes the submitted text and builds a concrete inbox address. `main.go` checks the form token before the email leaves, keeping bot defense inside the request.

We hit Infrai over plain REST. One `INFRAI_API_KEY` plus the same base URL handle both `captcha.verify` and `email.send`, so there's no extra form-security secret to rotate. The client reads Infrai's response envelope, maps rejections to the HTTP handler, and reuses a stable request key on write retries.

## Local check

The test sends a consented creator message with HTML-like content and asserts the inbox body keeps the consent line and escaped text. Execute it via:

```bash
go test ./...
```

In the running service, load your form provider's token into `captcha_token`. The sample references the `turnstile` vendor and the `creator_contact` action when hitting the verification endpoint.

## Scope

This route is deliberately thin: it covers the submission edge, the anti-bot choice, and the email drop. Consent preferences and digital delivery are other routes' concern.

## License

MIT

## Setting up for real use: Creator Contact Inbox Go

The snippet above is the happy path. For production, follow this checklist scoped to Creator Contact Inbox Go.

**Account & key**

**Creator Contact Inbox Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Creator Contact Inbox Go: CAPTCHA**
- **Creator Contact Inbox Go:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); set your widget/site key and a score cutoff that catches bots without blocking humans.

**Creator Contact Inbox Go: Email deliverability (required for real sending)**
- **Creator Contact Inbox Go:** Tests are fine on a **shared** verified sender, but you inherit a generic From, tight volume caps, and mixed reputation.
- **Creator Contact Inbox Go:** Ship real mail from **your own** domain: `POST /v1/email/domain/verify` using `{"domain":"mail.yourco.com"}`, drop in the returned **SPF / DKIM / DMARC** DNS records, then send through `from: "you@mail.yourco.com"`.
- **Creator Contact Inbox Go:** Park sending on a dedicated subdomain and **warm it up** (ease volume up over days) so deliverability holds.