# 0011 Structured logs, correlation IDs, OpenTelemetry opt-in

- Status: accepted
- Date: 2026-10-15
- Source: user

## Decision
- Logs: JSON to stdout. Go `log/slog`, Node `pino`. Required fields: `ts level msg service version env request_id trace_id`.
  Never log tokens, cookies, emails or magic-link URLs.
- Correlation:
  - Each service reads `X-Request-ID` (generating one if absent) and W3C `traceparent`.
  - Both are propagated on HTTP and gRPC metadata, and in event message metadata.
  - Traefik tracing (OTel) starts traces at the edge when enabled.
- OpenTelemetry SDK for traces and metrics in every backend. Export is enabled only when `OTEL_EXPORTER_OTLP_ENDPOINT` is set.
- Local: Compose profile `obs` runs `grafana/otel-lgtm`. Prod: no hosted stack yet. Docker `local` log driver with rotation.

## Revisit when
The first prod incident is hard to debug, or an SLO is defined.
