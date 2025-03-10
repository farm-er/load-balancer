<!-- omit in toc -->
# Load Balancer

<img src="./assets/images/load-balancer-vpc-svgrepo-com.svg" width=50% />

<!-- omit in toc -->
## Table of contents

- [Introduction](#introduction)
- [Installation](#installation)
- [Usage](#usage)
  - [Options](#options)
  - [File Format](#file-format)
- [Future](#future)
- [Sources](#sources)


## Introduction

- Load Balancers can help improve the system's availability and ressource management by deviding the load evenly to achieve a minimal response time.

- This project aims to support multiple algorithms out of the box, and a flexible design to let users add their own algorithms.

## Installation

```
  go install github.com/farm-er/load-balancer@latest
```

## Usage

### Options
```
  load-balancer 
    -name, -n [NAME_THE_SERVICE] DEFAULT load_balancer
    -strategy, -s [STRATEGY_OR_ALGORITHM_TO_USE] DEFAULT round-robin
    -file, -f [FILE_CONTAINING_URLS_POINTING_TO_SERVERS]
    -port, -p [SERVICE_PORT] DEFAULT 1234
```

**SEE load-balancer --help, -h FOR MORE INFORMATION**

### File Format
- current supported format

```
  http://127.0.0.1:8080
  http://127.0.0.1:8081
  http://127.0.0.1:8082
```

## Future

  - Support multiple input sources
  - Make the system more robust
  - Support for more strategies
  - Traffic replay option
  - Ability to add custom strategy that will be used by the load balancer

## Sources

- https://www.designgurus.io/course-play/grokking-system-design-fundamentals/doc/load-balancing-algorithms
- https://pkg.go.dev/net/http/httputil
