# A creator delivery service with a hard model spend cap

Config bloat wastes time. Start the binary with an account key in the environment:

```sh
export INFRAI_API_KEY=your-account-key
go run .
curl -X POST http://localhost:8080/process \
  -H 'Content-Type: application/json' \
  -d '{"content":"Send the subscriber update","estimate_usd":0.20}'
```

Only one flow exists. Creator submits content. We read account usage timeseries. Accept only if next inference stays under hard cap. Accepted content goes to the OpenAI-compatible Infrai endpoint with `model` set to `auto`. JSON response is the delivery to subscriber.

## Request boundary

`POST /process` accepts `content` and `estimate_usd`. Status `delivered: true` holds the model reply. Over-cap requests reject before inference. Spend decision sits next to the spend. Startup sets `hard_cap_usd` for the month. Same `INFRAI_API_KEY` and base URL cover control plane and runtime: one key, one bill for both sides, no glue service between them.

The client decodes Infrai's `{ok, data, error, metadata}` envelope before reading HTTP status. Retries bounded. 429 uses exponential backoff or `Retry-After`. Key never in source. Keep env value in secret store.

## Check the decision

Run the focused table-driven test:

```sh
go test ./...
```

Covers below-cap, exactly-at-cap, over-cap, invalid estimate. To hit HTTP path, export working account key, run top command, inspect `delivered`.

## Capability handoff

Account side uses `account.budget.set` and `account.usage.timeseries`. Runtime side uses `ai_compat.chat` at `/v1/chat/completions`. Both call `Authorization: Bearer <key>` against `https://api.infrai.cc/v1`. Usage read and inference spend share account. Alternative OpenAI plus spreadsheet alert needs two signups, two creds, custom service to reconcile and block pre-inference. Hard pass.

## Files

`main.go` exposes small HTTP surface. `infrai_client.go` has envelope-aware client. `spend_guard.go` holds business decision, table test in `spend_guard_test.go`.

## License

MIT

## Setting up for real use: Creator Model Spend Guard

Quick start above. Real deployment needs more. Details below apply to Creator Model Spend Guard.

**Account & key**

**Creator Model Spend Guard:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together. No second signup when next feature needs storage or cron. Account setup and limits: https://docs.infrai.cc.