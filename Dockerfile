FROM ubuntu:20.04

ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Install build env
RUN \
    apt-get update && \
    apt-get -y upgrade && \
    apt-get install -y build-essential python cmake git wget

# Install ragel
RUN \
    cd ~ && \
    wget http://www.colm.net/files/ragel/ragel-6.10.tar.gz && \
    tar -xvf ragel-6.10.tar.gz && \
    cd ragel-6.10 && \
    ./configure && \
    make && \
    make install && \
    ldconfig

# Install boost
RUN \
    cd ~ && \
    wget https://boostorg.jfrog.io/artifactory/main/release/1.69.0/source/boost_1_69_0.tar.gz && \
    tar -xvf boost_1_69_0.tar.gz && \
    cd boost_1_69_0 && \
    ./bootstrap.sh && \
    ./b2 --with-iostreams --with-random install && \
    ldconfig

# Install HyperScan
RUN \
    cd ~ && \
    git clone https://github.com/intel/hyperscan.git && \
    cd hyperscan && \
    mkdir cmake-build && \
    cd cmake-build && \
    cmake -DBUILD_SHARED_LIBS=on -DCMAKE_BUILD_TYPE=Release .. && \
    make -j8 && \
    make install && \
    ldconfig

# install go
RUN \
    apt-get install -y golang

# Copy code
WORKDIR /src
COPY . .