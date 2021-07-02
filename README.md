# Introduction 
For benchmarking the running time of HyperScan library under various regular expression and input data.

# Getting Started

## Requirement

Docker

## Build

```bash
docker build . -t hyperscan-test
```

## Run
```bash
# enter docker container as interactive mode
docker run --rm -it hyperscan-test /bin/bash
cd /src
go build -o bench
./bench
```

