# waflab-hyperscan-test

This repo is used for benchmarking the performance of HyperScan against the [ModSecurity Core Rule Set (v3.2/master)](https://github.com/coreruleset/coreruleset/tree/v3.2/master)

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

The output columns are:

1. Rule ID
2. Regex
3. Time for "Tough String"
4. Time for Random String
5. Current Progress

Usually we can use "4. Time for Random String" as the benchmarking time.
