FROM ubuntu:wily
MAINTAINER caktux

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update && \
    apt-get upgrade -q -y && \
    apt-get dist-upgrade -q -y && \
    apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 923F6CA9 && \
    echo "deb http://ppa.launchpad.net/shift/shift/ubuntu wily main" | tee -a /etc/apt/sources.list.d/shift.list && \
    apt-get update && \
    apt-get install -q -y geth

EXPOSE 54901
EXPOSE 54900

ENTRYPOINT ["/usr/bin/geth"]
