# Builder section
FROM golang:1.21-alpine3.18 as builder

WORKDIR /boltzmann

# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

# Setup deps modules
ADD go.mod go.sum /boltzmann/
RUN go mod download
RUN go mod verify
RUN go mod tidy -v

ADD . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /boltzmann-http ./cmd/http-server/main.go

# Runtime section
FROM alpine:3.18

ARG USER="boltzmann"

LABEL maintainer = "Boltzmann <docker@boltzmann.xyz>"
LABEL org.name="Boltzmann"
LABEL org.image.title="Bolztmann orchestrator service"

# Required by health checks
RUN apk --no-cache add curl

RUN adduser \
        -g "log daemon user" \
        --disabled-password \
        ${USER}

COPY --from=builder /boltzmann-http /usr/local/bin/boltzmann-http

EXPOSE 8080
RUN chown -R ${USER}:${USER} /var
RUN chmod 766 /var
USER ${USER}:${USER}

ENTRYPOINT ["/usr/local/bin/boltzmann-http"]