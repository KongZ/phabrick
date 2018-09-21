#################
## build stage ##
#################
FROM golang:1.10-alpine as builder
LABEL maintainer="KongZ <goonohc@gmail.com>"

RUN apk add --update --no-cache git

ADD . "/go/src/github.com/KongZ/phabrick"

WORKDIR /go/src/github.com/KongZ/phabrick/cmd
RUN go get ./...
RUN GO_ENABLED=0 GOOS=linux go build -o /app/phabrick .

############## ##
## final stage ##
#################
FROM alpine
LABEL maintainer="KongZ <goonohc@gmail.com>"

RUN apk add --update --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/phabrick /app/
COPY ./deployments/phabrick.yaml /app/phabrick.yaml
ENTRYPOINT [ "/app/phabrick" ]
EXPOSE 8000