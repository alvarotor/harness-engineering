# 0015 Asynchronous messaging via Redis pub/sub

- Status: deprecated
- Deprecated by: product choices deferred to Q7 (P-0016)
- Date: 2026-10-15
- Source: user (chosen over the RabbitMQ / Redis Streams recommendation)

## Decision
- Async communication between services uses **Redis pub/sub**. Channel naming:
  `<domain>.<event>.v<N>`. Payload: protobuf event from `contracts/proto/<domain>/events/v<N>/`,
  with metadata `event_id`, `occurred_at`, `request_id` and `traceparent`.
- Redis runs as one container per environment on the `data` network. It is added to Compose
  together with the first event contract (no unused infrastructure before that).
- Services publish and subscribe only through the `Publisher`/`Subscriber` ports in
  `internal/app`. The Redis adapter lives in `internal/adapters/messaging`.
- Consumers are idempotent (they deduplicate by `event_id`).

## Consequences (accepted by owner)
- Delivery is **at-most-once**: a message published while a subscriber is down, restarting or
  slow is lost. There is no retry, dead-letter queue or replay.
- Therefore pub/sub may carry only events whose loss is tolerable (notifications, cache
  invalidation, UI refresh). An OpenSpec change that adds an event must state why loss is tolerable.

## Revisit when
The first event whose loss is not tolerable appears. Upgrade path:
1. Redis Streams with consumer groups (same server, only the adapter changes).
2. RabbitMQ, if routing, dead-letter queues or managed AWS hosting are needed.
Either way, add a transactional outbox in the producer.
