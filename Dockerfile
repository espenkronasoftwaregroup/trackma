FROM golang:1.22-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o /trackma

EXPOSE 3100

CMD [ "/trackma" ]