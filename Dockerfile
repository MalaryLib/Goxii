FROM ubuntu:latest
WORKDIR /usr/src/app

RUN apt-get update
RUN apt-get install snapd
RUN snap install go --classic

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY src/ ./src/
COPY bin/ ./bin/

RUN ls -la .
