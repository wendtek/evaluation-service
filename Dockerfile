FROM golang:1.24.1-alpine

WORKDIR /app

# Copy everything and build
COPY . .
RUN go mod download && go build -o main .

EXPOSE 8080

CMD ["./main"] 
