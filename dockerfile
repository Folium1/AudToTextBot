FROM golang:1.20.5-alpine

WORKDIR /app
COPY go.mod /app
COPY go.sum /app
COPY . .

RUN go build -o main ./main.go
CMD [ "./main" ]