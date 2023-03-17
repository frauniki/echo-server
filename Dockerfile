FROM bufbuild/buf:latest as proto-gen

WORKDIR /src

COPY . .
RUN buf generate

FROM golang:1.20.2 as builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /src

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
COPY --from=proto-gen /src/gen ./gen

RUN go build -o echo-server ./cmd/echo-server/main.go

FROM gcr.io/distroless/static:latest

WORKDIR /

COPY --from=builder /src/echo-server /echo-server

USER nonroot

ENTRYPOINT [ "/echo-server" ]
