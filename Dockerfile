FROM golang:alpine

ENV APP_DIR $GOPATH/src/mongodm

RUN apk add --update --no-cache git && \
    mkdir -p $APP_DIR && \
    go get -u github.com/golang/dep/cmd/dep
    
ADD . $APP_DIR
    
WORKDIR $APP_DIR
