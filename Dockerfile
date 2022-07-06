FROM golang:1.16-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /go-app


FROM alpine:latest AS production
COPY --from=builder /go-app .
## we can then kick off our newly compiled
## binary exectuable!!
EXPOSE 9001

CMD [ "/go-app" ]