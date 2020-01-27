# Alkemi | Data Relay Service

![GitHub package.json version](https://img.shields.io/github/package-json/v/project-alkemi/alkemi-protocol) ![Built with Go](https://img.shields.io/badge/go-12.9-skyblue?style=flat&logo=go) [![Read the Docs](https://img.shields.io/badge/docs-yes-brightgreen.svg?logo=confluence)](https://alkemi.atlassian.net/wiki/spaces/ARCH/pages/93946093/Data+Relay+Service) [![The MIT License](https://img.shields.io/badge/license-MIT-orange.svg?style=flat-square&logo=read-the-docs)](http://opensource.org/licenses/MIT)

---

## Service Overview

The Data Relay Service represents a cluster of micro-services that only have one function. This function is to relay the trade completion information to the BookKeeper Nodes and the Price Index Service as required.

- This is done by using either rest or WebSockets to facilitate the data transfer. The BookKeeper Nodes will connect to the Data Relay Service to collect this information. Data Relay’s will be created as a micro-service that way we can have multiple data relays available for redundancy to the BookKeeper Nodes and global availability for Exchange Containers.

## Latest Deployment URLS

- [Data Relay Service API](https://relay.alkemi.tech/v1)
- [Data Relay Dashboard](https://relay.alkemi.tech)
  - User/Pass: admin/admin
- [Service Health](http://status.alkemi.tech/service/data-relay-service)

## Requirement

An application configuration file must be present in `/configs/config.json`. If this file is not present, application defaults to an auto-generated `/configs/dev.json` configuration file.

## Install

    go build bitbucket.org/project-alkemi/data-relay-service/cmd/data-relay-service

## Usage

    go run bitbucket.org/project-alkemi/data-relay-service/cmd/data-relay-service

## Run tests

    go test bitbucket.org/project-alkemi/data-relay-service/cmd/data-relay-service

## Author

👤 **Alkemi**

- Website: <https://alkemi.network>
- Twitter: [@AlkemiOfficial](https://twitter.com/AlkemiOfficial)
- Github: [@project-alkemi](https://github.com/project-alkemi)

## 🤝 Contributing

Contributions, issues and feature requests are welcome!

Please be sure to abide by our [Code Of Conduct](CODE_OF_CONDUCT.md)

Feel free to check [issues page](https://alkemi.atlassian.net/jira/software/projects/CD/boards/4).

If you come across an issue with the Data Relay Service, do a search in the [Issues](https://alkemi.atlassian.net/jira/software/projects/CD/boards/4) tab of this repo to make sure it hasn't been reported before. Follow these steps to help us prevent duplicate issues and unnecessary notifications going to the many people watching this repo:

- If the issue you found has been reported and is still open, and the details match your issue, give a "thumbs up" to the relevant posts in the issue thread to signal that you have the same issue. No further action is required on your part.
- If the issue you found has been reported and is still open, but the issue is missing some details, you can add a comment to the issue thread describing the additional details.
- If the issue you found has been reported but has been closed, you can comment on the closed issue thread and ask to have the issue reopened because you are still experiencing the issue. Alternatively, you can open a new issue, reference the closed issue by number or link, and state that you are still experiencing the issue. Provide any additional details in your post so we can better understand the issue and how to fix it.

## 📝 License

Copyright © 2020 [Alkemi Network](https://github.com/project-alkemi).

This project is [MIT](https://opensource.org/licenses/MIT) licensed.

    MIT License

    Copyright (c) 2020 Alkemi

    Permission is hereby granted, free of charge, to any person obtaining a copy
    of this software and associated documentation files (the "Software"), to deal
    in the Software without restriction, including without limitation the rights
    to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
    copies of the Software, and to permit persons to whom the Software is
    furnished to do so, subject to the following conditions:

    The above copyright notice and this permission notice shall be included in all
    copies or substantial portions of the Software.

    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
    IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
    FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
    AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
    LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
    OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
    SOFTWARE.
