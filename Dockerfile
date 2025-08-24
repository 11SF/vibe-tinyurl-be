############################
# STEP 1 build executable binary
############################
FROM golang:1.23.3-alpine as builder

# Install git, SSL CA certificates, and tzdata for timezone configuration.
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Set timezone to Thailand (Asia/Bangkok)
ENV TZ=Asia/Bangkok
RUN cp /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Create appuser
ENV USER=appuser
ENV UID=10001
# See https://stackoverflow.com/a/55757473/12429735RUN 
RUN adduser \    
  --disabled-password \    
  --gecos "" \    
  --home "/nonexistent" \    
  --shell "/sbin/nologin" \    
  --no-create-home \    
  --uid "${UID}" \    
  "${USER}"

WORKDIR $GOPATH/myapp
COPY . .
# Fetch dependencies.
# Using go mod with go 1.11
RUN go mod download
RUN go mod verify
# Build the binary
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/myapp ./cmd

############################
# STEP 2 build a small image
############################
FROM scratch

# Import from builder.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /usr/share/zoneinfo/Asia/Bangkok /usr/share/zoneinfo/Asia/Bangkok
COPY --from=builder /etc/localtime /etc/localtime
COPY --from=builder /etc/timezone /etc/timezone

# Copy our static executable
COPY --from=builder /go/bin/myapp /go/bin/myapp

# Use an unprivileged user.
USER appuser:appuser

# Run the hello binary.
ENTRYPOINT ["/go/bin/myapp"]

