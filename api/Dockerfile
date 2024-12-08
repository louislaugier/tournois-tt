FROM golang:1.23-alpine AS builder

WORKDIR /go/src/tournois-tt/api

COPY go.* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/api

FROM alpine:3.18
RUN apk --no-cache add ca-certificates

COPY --from=builder /go/bin/api /usr/local/bin/api

CMD ["api"] 