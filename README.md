# Project STAN

STAN is an extremely performant, lightweight reliable streaming platform built on [NATS](https://nats.io).

[Project Design Document](https://docs.google.com/document/d/1keDwK35YQnOXXKKy2HVV2oOnvEUPFyypT-Tplh8F89c/edit)

STAN provides the following high-level feature set.
- Log based.
- At-Least-Once Delivery model, giving reliable message delivery.
- Rate matched on a per subscription basis.
- Replay/Restart
- Last Value Semantics
