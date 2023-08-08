FROM ubuntu:latest
WORKDIR /usr/src/app

RUN apt-get update

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY src/ ./src/
COPY bin/ ./bin/

RUN ls -la .
