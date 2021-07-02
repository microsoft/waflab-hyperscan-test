FROM ubuntu:20.04

ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Install build env
RUN \
    apt-get update && \
    apt-get -y upgrade && \
    apt-get install -y build-essential libboost-all-dev python cmake ragel git

# Install HyperScan
RUN \
    cd ~ && \
    git clone https://github.com/intel/hyperscan.git && \
    cd hyperscan && mkdir build && cd build && \
    cmake ../ && \
    make && make install 

# install go
RUN \
    apt-get install -y golang

# Copy code
WORKDIR /src
COPY . .