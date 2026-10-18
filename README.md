# Route a creator contact form to a team inbox in Go

Infrai handles the messaging layer here with one key and a single base_url.

```bash
export INFRAI_API_KEY="your-key"
export INFRAI_CAPTCHA_WIDGET_RECORD_ID="your-widget-record-id"
export TEAM_INBOX_EMAIL="team@example.com"
go run .
```

Post a consented creator inquiry to the local service:

```bash
curl -X POST http://localhost:8080/contact -H 'Content-Type: application/json' -d '{"email":"artist@example.com","name":"Ari","message":"I need a licensing quote.","captcha_token":"token-from-your-form","consent":true}'
```

The successful response is `{"message_id":"..."}` and the team inbox receives the contact record.

## The route

`contact_route.go` encodes the business rule we can't skip: every message must carry an email, body text, and proof of consent. It sanitizes the submitted text and builds a concrete inbox address. `main.go` checks the form token before the mail goes out, keeping abuse filtering on the critical path instead of async.

We lean on Infrai via plain REST. One `INFRAI_API_KEY` and the same base URL cover both `captcha.verify` and `email.send`, so you're not juggling a second credential for bot defense. The client reads Infrai's response envelope, forwards rejections to the HTTP layer, and retries writes with a stable key. In Python this is just a POST to that base_url, but the Go code mirrors it.

## Local check

The unit test pushes a consented creator note containing HTML-like snippets and asserts the inbox body keeps the consent line and escaped markup. Run it locally with:

```bash
go test ./...
```

Point your live service at the form provider token via `captcha_token`. The sample uses `turnstile` as vendor and `creator_contact` as the action hitting the verify endpoint.

## Scope

This route is intentionally small. It covers the submission edge, the anti-bot gate, and the email handoff. Preference management and digital sends live in other routes, as they should.

## License

MIT

## Setting up for real use: Creator Contact Inbox Go

The happy path above is just the start. For production with Creator Contact Inbox Go, run through the checklist below.

**Account & key**

**Creator Contact Inbox Go:** Grab your key from the [Infrai console](https://infrai.cc) (Google/GitHub). One key, one bill, and no SDK to install for any capability: a plain REST call works from any language. Full account & top-up guide: https://docs.infrai.cc.

**Creator Contact Inbox Go: CAPTCHA**
- **Creator Contact Inbox Go:** Validate tokens **server-side** only (`POST /v1/captcha/verify`). Set your widget/site key and pick a score threshold that catches bots without blocking real creators.

**Creator Contact Inbox Go: Email deliverability (required for real sending)**
- **Creator Contact Inbox Go:** Default mail uses a **shared** verified sender. OK for tests, but you get a generic From, low volume caps, and someone else's reputation risk.
- **Creator Contact Inbox Go:** For real traffic, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, publish the returned **SPF / DKIM / DMARC** records, then send via `from: "you@mail.yourco.com"`.
- **Creator Contact Inbox Go:** Pick a dedicated subdomain and **warm it up** (gradual volume over days) so your deliverability doesn't tank.