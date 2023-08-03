FROM ubuntu:16.04

WORKDIR /usr/src/app

RUN apt-get update
RUN apt-get install -y vim

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY . .
