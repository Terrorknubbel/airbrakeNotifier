FROM golang:latest AS builder

RUN apt-get update && apt-get install -y ca-certificates openssl

ARG cert_location=/usr/local/share/ca-certificates

# Get certificate from "https://api.airbrake.io/api/v4"
RUN openssl s_client -showcerts -connect api.airbrake.io:443 </dev/null 2>/dev/null | sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p' | openssl x509 -outform PEM > ${cert_location}/airbrake.crt


# Get certificate from "proxy.golang.org"
RUN openssl s_client -showcerts -connect proxy.golang.org:443 </dev/null 2>/dev/null | sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p' | openssl x509 -outform PEM >  ${cert_location}/proxy.golang.crt

# Update certificates
RUN update-ca-certificates

WORKDIR /app
COPY . /app

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN  GO111MODULE="on" CGO_ENABLED=0 GOOS=linux go build -o /usr/local/bin/airbrakeNotifier main.go

FROM alpine:latest
LABEL maintainer="Terrorknubbel"
RUN apk add --no-cache bash
WORKDIR /app
COPY --from=builder /usr/local/bin/airbrakeNotifier /usr/bin/
CMD ["airbrakeNotifier"]
