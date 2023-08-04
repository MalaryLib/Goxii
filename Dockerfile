FROM golang:1.17.6-alpine as Build
WORKDIR /usr/src/app


# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY src/ ./src/
COPY bin/ ./bin/

RUN ls -la .
