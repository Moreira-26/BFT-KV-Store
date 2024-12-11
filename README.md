# Byzantine Fault Tolerat Key-Value Store

This project was built as the final project for the 'Second ACM Europe Summer School on
Distributed and Replicated Environments (DARE 2024)'

## Requirements:

The Go programming language version 1.23.1.

## How to build and run

### Build

```bash
go build bftkvstore.go
```

### Run

```bash
./bftkvstore --port <port> --config <config-folder>
```

The configuration folder can be omitted, a new one will be generated on startup.

#### Connecting nodes

To connect two nodes run the following command:

```bash
bash cmds/connect.sh <node1-address> <node1-port> <node2-address> <node2-port>
```

