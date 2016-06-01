# adjusted from https://github.com/shift/cpp-shift/blob/develop/docker/Dockerfile
FROM ubuntu:14.04

ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update
RUN apt-get upgrade -y

# Shift dependencies
RUN apt-get install -qy build-essential g++-4.8 git cmake libboost-all-dev libcurl4-openssl-dev wget
RUN apt-get install -qy automake unzip libgmp-dev libtool libleveldb-dev yasm libminiupnpc-dev libreadline-dev scons
RUN apt-get install -qy libjsoncpp-dev libargtable2-dev

# NCurses based GUI (not optional though for a succesful compilation, see https://github.com/shift/cpp-shift/issues/452 )
RUN apt-get install -qy libncurses5-dev

# Qt-based GUI
# RUN apt-get install -qy qtbase5-dev qt5-default qtdeclarative5-dev libqt5webkit5-dev

RUN sudo apt-get -y install software-properties-common

# LLVM-3.5
RUN wget -O - http://llvm.org/apt/llvm-snapshot.gpg.key|sudo apt-key add -
RUN echo "deb http://llvm.org/apt/trusty/ llvm-toolchain-trusty-3.5 main\ndeb-src http://llvm.org/apt/trusty/ llvm-toolchain-trusty-3.5 main" > /etc/apt/sources.list.d/llvm-trusty.list
RUN apt-get update
RUN apt-get install -qy llvm-3.5 libedit-dev

# Fix llvm-3.5 cmake paths
RUN mkdir -p /usr/lib/llvm-3.5/share/llvm && ln -s /usr/share/llvm-3.5/cmake /usr/lib/llvm-3.5/share/llvm/cmake


# Shift PPA
RUN apt-get install -qy software-properties-common
RUN add-apt-repository ppa:shift/shift
RUN apt-get update
RUN apt-get install -qy libcryptopp-dev libjson-rpc-cpp-dev

# this is a workaround, to make sure that docker's cache is invalidated whenever the git repo changes
ADD https://api.github.com/repos/shift/cpp-shift/git/refs/heads/develop unused.txt 

# Build Shift (HEADLESS)
RUN git clone --depth=1 --branch develop https://github.com/shift/cpp-shift
RUN mkdir -p cpp-shift/build
RUN cd cpp-shift/build && cmake .. -DCMAKE_BUILD_TYPE=Debug -DVMTRACE=1 -DPARANOIA=1 -DEVMJIT=1 && make -j $(cat /proc/cpuinfo | grep processor | wc -l) && make install
RUN ldconfig

ENTRYPOINT ["/cpp-shift/build/test/checkRandomStateTest"]
