FROM golang:1.9-alpine3.6

RUN apk add --no-cache --update git && \
    go get -u github.com/golang/dep/cmd/dep
RUN git clone https://github.com/nicolai86/couchdb-operator /go/src/github.com/nicolai86/couchdb-operator && \
    cd /go/src/github.com/nicolai86/couchdb-operator && \
    dep ensure && \
    go install github.com/nicolai86/couchdb-operator

FROM alpine:3.6

RUN apk --no-cache --update add ca-certificates && update-ca-certificates

COPY --from=0 /go/bin/couchdb-operator /

ENTRYPOINT ["/couchdb-operator"]
