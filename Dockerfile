FROM golang:alpine3.15 AS build

WORKDIR /usr/fairos
COPY go.mod go.sum /usr/fairos/
RUN go mod download
COPY . /usr/fairos/
RUN apk add --update --no-cache make=4.3-r0 gcc=10.3.1_git20211027-r0 git=2.34.2-r0 musl-dev=1.2.2-r7 libc-dev=0.7.2-r3 linux-headers=5.10.41-r0 bash
RUN make binary

FROM alpine:3.15

ARG CONFIG
ENV CONFIG=$CONFIG

RUN addgroup -g 10000 fds
RUN adduser -u 10000 -G fds -h /home/fds -D fds
USER fds
RUN if [ -n "$CONFIG" ]; then  echo -e "$CONFIG" > ~/.dfs.yaml; fi
EXPOSE 9090

COPY --from=build /usr/fairos/dist/dfs /usr/local/bin/dfs
COPY --from=build /usr/fairos/dist/dfs-cli /usr/local/bin/dfs-cli

ENTRYPOINT ["dfs"]
