# Introduction 
For benchmarking the running time of HyperScan library under various regular expression and input data.

# Getting Started

## Requirement

More details requirement, please refer to the Dockerfile 

## Build

```bash
g++ hyperscan.c -o hyperscan -lhs
```

## Run
```bash
./hyperscan <inputfile> <regular expression>
```

**Example**

```bash
./hyperscan generate.txt regex.txt
```

