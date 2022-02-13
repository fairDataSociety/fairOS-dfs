FROM golang:1.17-alpine AS build

WORKDIR /usr/fairos
COPY go.mod go.sum /usr/fairos/
RUN go mod download
COPY . /usr/fairos/
RUN apk add --update make gcc git musl-dev libc-dev linux-headers
RUN make binary

FROM alpine:3.15

RUN addgroup -g 10000 fds
RUN adduser -u 10000 -G fds -h /home/fds -D fds
USER fds
EXPOSE 9090

COPY --from=build /usr/fairos/dist/dfs /usr/local/bin/dfs
COPY --from=build /usr/fairos/dist/dfs-cli /usr/local/bin/dfs-cli

ENTRYPOINT ["dfs"]
