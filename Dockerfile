FROM golang:1.12.7 as builder

RUN apt-get update && apt-get install ca-certificates

WORKDIR /src

COPY go.mod /src/go.mod
COPY go.sum /src/go.sum

RUN go mod download

# Perform the build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /dist/argocd-notifications-controller ./cmd/main.go

FROM debian

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /dist/argocd-notifications-controller /app/argocd-notifications-controller
COPY --from=builder /src/assets/config.yaml /app/assets/config.yaml
