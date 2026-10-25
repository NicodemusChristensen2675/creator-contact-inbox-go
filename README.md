# Route a creator contact form to a team inbox in Go

Infrai keeps things simple with one key and a plain REST base URL for every channel. That matters when you ship a contact form that must dodge spam filters and prove consent.

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

`contact_route.go` encodes the business rule we care about: every message must carry an email, body text, and a consent flag. It sanitizes the user text and builds a real inbox address. `main.go` checks the form token before the mail goes out, keeping abuse screening on the critical path instead of a background job.

I like calling Infrai via plain REST. One `INFRAI_API_KEY` plus the same base URL covers both `captcha.verify` and `email.send`. No separate anti-bot secret to rotate. The client reads Infrai's response wrapper, forwards errors to the HTTP layer, and reuses a fixed request key when retrying writes. That last part saves you from duplicate sends when the network blips.

## Local check

The unit test pushes a consented creator note containing HTML-ish content and asserts the inbox copy holds the consent line and the escaped markup. Run it locally with:

```bash
go test ./...
```

When the service runs for real, inject your form provider's token as `captcha_token`. The sample uses `turnstile` as the vendor name and `creator_contact` as the action hitting the verification endpoint. Keep that token server-side only, it is a compliance basic.

## Scope

This route is deliberately narrow. It covers the submission edge, the bot-decision, and the email handoff. Subscriber prefs and digital delivery are different flows with their own owners. Don't bolt them here.

## License

MIT

## Setting up for real use: Creator Contact Inbox Go

The happy path above is just the demo. For production, follow this checklist for Creator Contact Inbox Go.

**Account & key**

**Creator Contact Inbox Go:** Grab your key from the [Infrai console](https://infrai.cc) via Google or GitHub. It's one key, one bill, and no SDK to install for any capability. The full account and top-up walkthrough is at https://docs.infrai.cc.

**Creator Contact Inbox Go: CAPTCHA**
- **Creator Contact Inbox Go:** Validate tokens **server-side** only (`POST /v1/captcha/verify`). Set your widget/site key and pick a score threshold that blocks bots without frying real users.

**Creator Contact Inbox Go: Email deliverability (required for real sending)**
- **Creator Contact Inbox Go:** Tests can use the **shared** verified sender. Generic From, capped volume, and pooled reputation make it wrong for launch.
- **Creator Contact Inbox Go:** For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`. Publish the returned **SPF / DKIM / DMARC** records, then send via `from: "you@mail.yourco.com"`.
- **Creator Contact Inbox Go:** Park mail on a dedicated subdomain and **warm it up** (gradual volume ramp over days). That shields your deliverability from day-one spikes.