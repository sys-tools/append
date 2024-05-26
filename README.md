## Append

An embedded append-only storage system built around append-only logs.

## Table of Contents

- [Append](#append)
- [Table of Contents](#table-of-contents)
- [Introduction](#introduction)
- [Features](#features)
- [Installation](#installation)


## Introduction

Append is an embedded storage system built around immutable append-only logs. It is designed to be simple, fast, and reliable.
Append is built to provide the core data storage functionality for building distributed data systems like Kafka, Pulsar and Nats. It's not meant to be a fully featured system, thus, it does't provide features like replication, sharding, consumer groups, consistency guraantees etc. It's meant to be a building block for building such systems.

Append is meant to be part of a larger ecosystem. 

## Features

- [x] Append-only logs.
- [ ] Support multiple data formats:
  - [x] JSON
  - [ ] Avro
  - [ ] String
- [x] Provide essential algorithms like `seek`, `count`, `read`, `write`, `watch` etc.
- [ ] Support for multiple log files.
- [ ] Support retention algorithms.
  - [ ] Time-based retention.
  - [ ] Size-based retention.
  - [ ] Custom retention.
- [ ] Support for multi-level logging.
- [ ] Support exported metrics.
- [ ] Support for zero-copy reads.


## Installation

```bash
go get github.com/append/append
```