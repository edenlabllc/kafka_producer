FROM golang:1.11.5-alpine3.8 as builder

RUN apk add --virtual .build-deps \
    alpine-sdk \
    cmake \
    libssh2 libssh2-dev\
    git \
    xz

WORKDIR /src

ENV GO111MODULE=on

ADD . .

ARG APP_NAME

RUN go get -d -v .

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o ${APP_NAME}_build main.go
# -mod=vendor -a -installsuffix cgo

RUN strip --strip-unneeded ${APP_NAME}_build

FROM erlang:21.2.5-alpine

RUN apk update && apk add --no-cache \
  ncurses-libs \
  zlib \
  ca-certificates \
  openssl \
  bash

RUN rm -rf /var/cache/apk/*

WORKDIR /root

ARG APP_NAME

COPY --from=builder /src/${APP_NAME}_build .

CMD ["./${APP_NAME}_build"]