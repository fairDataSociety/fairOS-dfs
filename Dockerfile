FROM golang:1.21.4-alpine AS build

WORKDIR /usr/fairos
COPY go.mod go.sum /usr/fairos/
RUN go mod download
COPY . /usr/fairos/

#skipcq: DOK-DL3018
RUN apk add --update --no-cache make gcc git musl-dev libc-dev linux-headers bash \
    && make binary

FROM alpine:3.18

ARG CONFIG
ENV CONFIG=$CONFIG

RUN addgroup -g 10000 fds \
    && adduser -u 10000 -G fds -h /home/fds -D fds
USER fds
RUN if [ -n "$CONFIG" ]; then  echo -e "$CONFIG" > ~/.dfs.yaml; fi
EXPOSE 9090

COPY --from=build /usr/fairos/dist/dfs /usr/local/bin/dfs
COPY --from=build /usr/fairos/dist/dfs-cli /usr/local/bin/dfs-cli

ENTRYPOINT ["dfs"]
