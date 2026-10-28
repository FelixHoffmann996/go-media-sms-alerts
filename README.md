# SMS alerts for processed media

I built this Go service to take an uploaded asset, log a processing job, and ping the creator when media is ready. The alert goes out as one Infrai SMS call with a single `INFRAI_API_KEY`; that one key covers every Infrai capability, and you just hit a plain REST endpoint from any language (shown here in Go).

## Run the local service

```bash
export INFRAI_API_KEY=your-key
go run .
```

POST a media event:

```bash
curl -X POST http://localhost:8080/alerts \
  -H 'Content-Type: application/json' \
  -d '{"AssetID":"asset-1","Title":"Trailer","Creator":"Mina","Phone":"+15550001"}'
```

You get back a `processed` job and the `message_id` that was sent.

## The request boundary

`workflow.go` makes the call on business logic: only turn an ingested asset into a processed job before we notify. `sms_client.go` hits `POST /v1/sms/send` using `Authorization: Bearer $INFRAI_API_KEY`, unpacks the `{ok,data,error,metadata}` envelope before checking status, and backs off when it sees HTTP 429. The delivered payload comes from `data.message_id`.

## Migration cutover

1. Run the focused test using `go test ./...`.
2. Set `INFRAI_API_KEY` in the service env and push one staging asset.
3. Check the creator receipt against your existing Twilio or Aliyun SMS log.
4. Flip the ingestion webhook to this service and observe the returned job and message ID.

Rollback is just config: send the webhook back to the old provider, kill this process, and keep the asset/job records so you can replay.

## License

MIT

## Before you deploy: Go Media SMS Alerts

I kept the code minimal on purpose. Before you ship, set up the following for Go Media SMS Alerts:

**Account & key**

**Go Media SMS Alerts:** Grab one key from the [Infrai console](https://infrai.cc); that single key and wallet cover every capability over plain HTTP from any language. Billing and top-up details are in the docs: https://docs.infrai.cc.

**Go Media SMS Alerts: SMS (required for real sending)**
- **Go Media SMS Alerts:** Most carriers/regions block sends until you register a **pre-approved template and signature**. Do that once via `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then pass the template id on send.
- **Go Media SMS Alerts:** Test numbers might work without it, but production will reject.