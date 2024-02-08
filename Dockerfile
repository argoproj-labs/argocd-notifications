ARG BUILDPLATFORM=linux/amd64
FROM --platform=$BUILDPLATFORM golang:1.16.2 as builder

RUN apt-get update && apt-get install ca-certificates

WORKDIR /src

ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM
ARG BUILDPLATFORM

COPY go.mod /src/go.mod
COPY go.sum /src/go.sum

RUN go mod download

# Perform the build
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -o /app/argocd-notifications ./cmd
RUN ln -s /app/argocd-notifications /app/argocd-notifications-backend

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/argocd-notifications /app/argocd-notifications
COPY --from=builder /app/argocd-notifications-backend /app/argocd-notifications-backend

# User numeric user so that kubernetes can assert that the user id isn't root (0).
# We are also using the root group (the 0 in 1000:0), it doesn't have any
# privileges, as opposed to the root user.
USER 1000:0
