# ==== Build Stage =====================================================================================================
FROM golang:1.24-bookworm AS builder

RUN mkdir -p /app/bin
WORKDIR /app
COPY go.mod go.sum Makefile .
RUN go mod download

COPY cmd /app/cmd
COPY internal /app/internal
COPY .git /app/.git

RUN make app/build CGO_ENABLED=0 GOOS=linux GOARCH=amd64


# ==== Run Stage =======================================================================================================
FROM debian:12.10-slim AS runner

RUN useradd -ms /bin/bash -u 1001 app
USER app

WORKDIR /app

COPY --from=builder --chown=app:app /app/bin/the-hunt-api.linux.amd64 .

CMD ["/app/the-hunt-api.linux.amd64"]
