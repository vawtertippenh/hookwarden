# hookwarden

Lightweight webhook receiver and inspector with replay, logging, and HMAC signature validation.

---

## Installation

```bash
go install github.com/yourusername/hookwarden@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/hookwarden.git && cd hookwarden && go build ./...
```

---

## Usage

Start the webhook receiver on a given port and secret:

```bash
hookwarden --port 8080 --secret "mysecret" --log ./hooks.log
```

hookwarden will listen for incoming POST requests, validate the `X-Hub-Signature-256` HMAC header, log the full request payload, and expose a simple inspector UI at `http://localhost:8080/_inspector`.

**Replay a captured request:**

```bash
hookwarden replay --id <request-id> --target https://your-service.example.com/webhook
```

**Example config file (`hookwarden.yaml`):**

```yaml
port: 8080
secret: mysecret
log: ./hooks.log
inspector: true
```

---

## Features

- HMAC SHA-256 signature validation
- Full request logging (headers + body)
- Web-based inspector UI
- One-command request replay
- Zero external dependencies

---

## License

MIT © 2024 hookwarden contributors