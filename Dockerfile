FROM golang:1.8


COPY . /go/src/github.com/play-with-docker/api.play-with-moby.com

WORKDIR /go/src/github.com/play-with-docker/api.play-with-moby.com

RUN go get -v -d ./...

RUN CGO_ENABLED=0 go build -a -installsuffix nocgo -o /go/bin/api .


FROM alpine

RUN apk --update add ca-certificates
RUN mkdir -p /app/pwm

COPY --from=0 /go/bin/api /app/api

WORKDIR /app
CMD ["./api"]

EXPOSE 8080
