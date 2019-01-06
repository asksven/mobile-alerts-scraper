# We want to smallest and most secure image possible
# run under non-root
# layer the image an build from "scratch"

# Start with golang:alpine for building
FROM golang:alpine as builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Create appuser
RUN adduser -D -g '' appuser

COPY . $GOPATH/src/github.com/asksven/mobile-alerts-scraper
WORKDIR $GOPATH/src/github.com/asksven/mobile-alerts-scraper

# Fetch dependencies.
# Using go get.
RUN go get -d -v
# Build the binary.
# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/mobile-alerts-scraper

# Use the smallest possible image as we don't need much to run
FROM scratch

# Import from builder.
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

COPY --from=builder /go/bin/mobile-alerts-scraper /go/bin/mobile-alerts-scraper

# Use an unprivileged user.
USER appuser

CMD ["/go/bin/mobile-alerts-scraper"]