FROM golang:1.16 AS build

WORKDIR /usr/fairos
COPY go.mod go.sum /usr/fairos/
RUN go mod download
COPY . /usr/fairos/
RUN make binary

FROM debian:10.9-slim

EXPOSE 9090

COPY --from=build /usr/fairos/dist/dfs /usr/local/bin/dfs
COPY --from=build /usr/fairos/dist/dfs-cli /usr/local/bin/dfs-cli

ENTRYPOINT ["dfs"]
