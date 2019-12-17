# Data Relay Service
[![Documentation](https://img.shields.io/badge/documentation-yes-brightgreen.svg)](https://alkemi.atlassian.net/wiki/spaces/EI/pages/2949225/Data+Relay)

> The DR ( Data Relay) only has one function. This function is to relay the trade completion information to the currently completed oracles and the VWAP index service. This is done by using either rest or WebSockets to facilitate the data transfer. The Oracles will connect to the DR’s to collect this information. DR’s will be created as a micro-service that way we can have multiple data relays available for redundancy to the oracles.

## Install

```sh
go build bitbucket.org/project-alkemi/data-relay-service/cmd/data-relay-service
```

## Usage

```sh
go run bitbucket.org/project-alkemi/data-relay-service/cmd/data-relay-service
```

## Run tests

```sh
go test bitbucket.org/project-alkemi/data-relay-service/cmd/data-relay-service
```


## 🤝 Contributing

Contributions, issues and feature requests are welcome!

Feel free to check [issues page](https://alkemi.atlassian.net/jira/software/projects/CD/boards/4).
