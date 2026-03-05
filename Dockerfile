FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o scheduler ./cmd/scheduler

FROM alpine:3.21

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/scheduler .

USER appuser

ENTRYPOINT [ "./scheduler" ]
