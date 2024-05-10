FROM golang:1.22-alpine AS BuildStage

WORKDIR /

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -ldflags "-s -w" -o /trackma

FROM alpine:latest

WORKDIR /

COPY --from=BuildStage /trackma /trackma
COPY ./migrations/ ./migrations/
COPY ./public/ ./public/
COPY ./views/ ./views/
COPY ./dbip-country-lite.csv ./

EXPOSE 3100

CMD [ "/trackma" ]