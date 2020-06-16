FROM golang:1.13.6 as builder

RUN apt-get update && apt-get install ca-certificates

WORKDIR /src

COPY go.mod /src/go.mod
COPY go.sum /src/go.sum

RUN go mod download

# Perform the build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /dist/argocd-notifications ./cmd

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /dist/argocd-notifications /app/argocd-notifications

# User numeric user so that kubernetes can assert that the user id isn't root (0).
# We are also using the root group (the 0 in 1000:0), it doesn't have any
# privileges, as opposed to the root user.
USER 1000:0
