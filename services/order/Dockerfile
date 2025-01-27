FROM golang:1.22-alpine3.19
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o main .
EXPOSE 8001
CMD ["./main"]