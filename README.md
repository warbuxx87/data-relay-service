# Data Relay Service

> Go service that relays completed-trade data from exchanges to oracles and a VWAP index over REST and WebSockets, with TLS and Redis.

<p>
<img src="https://img.shields.io/badge/Go-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
<img src="https://img.shields.io/badge/Redis-DC382D?style=flat-square&logo=redis&logoColor=white" alt="Redis" />
<img src="https://img.shields.io/badge/MySQL-4479A1?style=flat-square&logo=mysql&logoColor=white" alt="MySQL" />
<img src="https://img.shields.io/badge/Swagger-85EA2D?style=flat-square&logo=swagger&logoColor=black" alt="Swagger" />
<img src="https://img.shields.io/badge/Sentry-362D59?style=flat-square&logo=sentry&logoColor=white" alt="Sentry" />
</p>

The relay layer of the platform. Exchange containers push completed trades to a Data Relay (DR), and the DR fans that data out to subscribed oracles and the VWAP [price-index-service](https://github.com/warbuxx87/price-index-service) over REST or WebSockets. Relays register themselves with the [discovery service](https://github.com/warbuxx87/dr-discovery-service) so oracles can find them, and they authenticate through the [switchboard](https://github.com/warbuxx87/switch-board-service).

## Where it fits

Part of Alkemi's distributed market-data platform: a set of Go microservices that collect completed trades from exchange containers, relay them to oracles and a VWAP price index, and settle results on-chain.

```mermaid
flowchart LR
	EX[Exchange containers] -->|completed trades| DR[Data Relay Service]
	SB[Switchboard Service] -. whitelist + service keys .-> DR
	SB -. service keys .-> PI
	DS[Discovery Service] -. relay registry .-> OR
	DR -->|REST / WebSocket| PI[Price Index Service<br/>VWAP]
	DR -->|REST / WebSocket| OR[Oracles / Book-keeper nodes]
	OR -->|settlement votes| SC[(Settlement contract)]
	ADM[Switchboard Dashboard] --> SB
```

## How it works

- `cmd/data-relay-service` runs the TLS-enabled Echo server, and `cmd/data-relay-service-dev-client` is a test client.
- `internal/app/data-relay-service/managers` handles subscriber management and fan-out, using Redis for pub/sub state.
- Shared message types come from Alkemi's common message package.
- Errors are reported to Sentry and logs go through logrus.

## Stack

Go, Redis, MySQL, Swagger, Sentry

## Running locally

```bash
cp configs/config.json.example configs/config.json
go run ./cmd/data-relay-service
```

## Credits

Built at Alkemi by Raynaldo Rivera and Curtis Goolsby (2019), as part of the platform I founded and led.

---
<sub>Part of <a href="https://github.com/warbuxx87">Ryan Breen's</a> work · <a href="https://www.linkedin.com/in/breen-ryan/">LinkedIn</a></sub>
