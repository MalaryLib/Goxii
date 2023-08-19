FROM ubuntu:latest

WORKDIR /usr/share/app

RUN apt-get update
RUN apt-get install -y libpcap-dev

COPY ./Goxii .
COPY ./templates ./templates
COPY ./configuration.json .
COPY ./descendent.json .
