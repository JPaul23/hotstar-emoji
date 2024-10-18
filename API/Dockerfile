# Distroless Dockerfile
FROM golang:1.23.2-bullseye AS base

RUN adduser \
  --disabled-password \
  --gecos "" \
  --home "/nonexistent" \
  --shell "/sbin/nologin" \
  --no-create-home \
  --uid 65532 \
  appuser

WORKDIR $GOPATH/src/hotstar-emoji-api/app/

COPY .env ./

COPY . .

RUN go mod download
RUN go mod verify

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /main .

# FROM gcr.io/distroless/static-debian11
FROM alpine:latest

RUN adduser -D -u 65532 appuser

COPY --from=base /main .

USER appuser:appuser

# Expose port 3000 to the outside world
EXPOSE 3000

CMD ["./main"]