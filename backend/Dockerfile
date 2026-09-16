FROM golang:1.27-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/feedback-investigator-api \
    ./server.go

FROM alpine:3.22

RUN addgroup -S appgroup \
    && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder \
    /out/feedback-investigator-api \
    /app/feedback-investigator-api

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/feedback-investigator-api"]