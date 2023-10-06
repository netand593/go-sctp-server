# SCTP server

A go program packed into a docker container to deploy an sctp-server.

## Build

```bash
docker build -t go-sctp-server .
```

## Run

```bash
docker run -p 38412:38412 -p go-sctp-server
```

This can be tested using `ncat` or `socat`:

```bash
ncat --sctp localhost 38412
```
