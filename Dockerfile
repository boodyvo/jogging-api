FROM golang:1.14-alpine

# TODO(boodyvo): Add mod install

RUN apk add --update git alpine-sdk
ENV GO111MODULE=off
RUN go get -v github.com/oxequa/realize
ENV GO111MODULE=on
#COPY . /app
#WORKDIR /app/services/api/cmd/api-service/
#RUN go install .
#WORKDIR /app/services/gateway/cmd/gateway-service/
#RUN go install .
WORKDIR /app
COPY ./docs /docs

CMD make run-dev