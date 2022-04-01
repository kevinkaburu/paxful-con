FROM golang:1.14-alpine AS builder

RUN set -ex &&\
    apk add --no-progress --no-cache \
      gcc \
      musl-dev

# Set necessary environmet variables needed for our image

ENV GO111MODULE=on \
    # CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory /build

WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go get -d -v
#RUN go build -a -tags musl -installsuffix cgo -ldflags '-extldflags "-static"'  -o main main.go
RUN go build -o main main.go


# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy binary from build to main folder
RUN cp /build/main .

# Build a small image
FROM alpine

COPY --from=builder /dist/main /

RUN apk add --no-cache tzdata

ENV TZ=Africa/Nairobi
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

COPY deployment/docker.env .env


ENV ENV prod
EXPOSE 51591
# Command to run
ENTRYPOINT ["/main"]