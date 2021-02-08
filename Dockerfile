FROM golang:alpine AS builder

WORKDIR /opt/app

COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy source code and build
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o ./bin/go-migrate

#
# Scratch docker image containing nothing but the application
#
FROM alpine AS production
# Start at /opt/app
WORKDIR /opt/app
# Copy our static executable.
COPY --from=builder /opt/app/bin/go-migrate /opt/app/go-migrate
CMD ["/opt/app/go-migrate"]
