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

# Install PCRE2 
RUN \
    apt-get install -y libpcre2-dev

# Install RE2 & CRE2
RUN LD_LIBRARY_PATH="/usr/local/lib"
RUN export LD_LIBRARY_PATH
RUN \
    cd ~ && \
    git clone https://github.com/google/re2.git && \
    cd re2 && make && make install 
RUN \
    cd ~ && \
    apt install -y pkg-config texinfo && \
    git clone https://github.com/marcomaggi/cre2.git && \
    cd cre2 && \
    sh autogen.sh && \
    mkdir build && cd build && ../configure --enable-maintainer-mode && make && make install
    
