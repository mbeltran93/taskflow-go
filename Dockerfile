# --- build stage -----------------------------------------------------------
FROM golang:1.26-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 gives a fully static binary with no libc dependency, which is
# what lets the final image be `scratch` - no shell, no package manager, no
# OS at all, just the Go binary. This is one of Go's real advantages over a
# JVM stack for a cloud-native image: there is no runtime to ship.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/taskflow-go ./cmd/api

# --- runtime stage ----------------------------------------------------------
FROM scratch

# TLS root certificates, in case the binary ever needs to make an outbound
# HTTPS call; harmless to include and keeps the image ready for that.
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/taskflow-go /taskflow-go

EXPOSE 8080
ENTRYPOINT ["/taskflow-go"]
