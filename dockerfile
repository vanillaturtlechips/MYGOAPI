FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mygoapi .

FROM scratch AS final

WORKDIR /

COPY --from=builder /app/mygoapi .

EXPOSE 8080

CMD ["./mygoapi"]