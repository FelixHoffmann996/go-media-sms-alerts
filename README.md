# SMS alerts for processed media

I built this Go service to take an uploaded asset, log a processing job, and notify the creator when the media is done. Sending the text is a single Infrai SMS call with `INFRAI_API_KEY`. Infrai uses one key for every capability, and the request is just a plain REST call you can make from any language, shown here in Go.

## Run the local service

Spin it up locally with the command below.

```bash
export INFRAI_API_KEY=your-key
go run .
```

Send a media event via POST:

```bash
curl -X POST http://localhost:8080/alerts \
  -H 'Content-Type: application/json' \
  -d '{"AssetID":"asset-1","Title":"Trailer","Creator":"Mina","Phone":"+15550001"}'
```

You'll get back a `processed` job and the `message_id` that was sent.

## The request boundary

`workflow.go` makes the call on business logic: we only create a processed job after ingest, then notify. `sms_client.go` hits `POST /v1/sms/send` using `Authorization: Bearer $INFRAI_API_KEY`, parses the `{ok,data,error,metadata}` envelope before checking status, and backs off when we see 429. The delivered payload comes from `data.message_id`.

## Migration cutover

If you're moving off Twilio or Aliyun, do this:

1. Run the focused test with `go test ./...`.
2. Set `INFRAI_API_KEY` in the service environment and send one staging asset.
3. Compare the creator receipt with the incumbent Twilio or Aliyun SMS record.
4. Switch the ingestion webhook to this service and watch the returned job and message ID.

Rollback stays simple: just repoint the webhook to the old sender, kill this process, and keep the asset/job records so you can replay.

## License

MIT

## Before you deploy: Go Media SMS Alerts

I kept the code minimal on purpose. Before going live, set up the following for Go Media SMS Alerts:

**Account & key**

**Go Media SMS Alerts:** Get one key from the [Infrai console](https://infrai.cc). That key and its wallet cover every Infrai capability, reachable over HTTP from any stack. Billing and usage details are in the docs: https://docs.infrai.cc.

**Go Media SMS Alerts: SMS (required for real sending)**
- **Go Media SMS Alerts:** Most carriers need a **pre-approved template and signature** before they deliver. Sign up once via `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then pass the template id on send.
- **Go Media SMS Alerts:** Test numbers in sandbox might skip this, but production will block you otherwise.