FROM golang:1.5.1

WORKDIR /go/src/github.com/asiainfoLDP/datafoundry_oauth2
ADD . /go/src/github.com/asiainfoLDP/datafoundry_oauth2

EXPOSE 9443

ENV SERVICE_NAME=datafoundry_oauth2

RUN GO15VENDOREXPERIMENT=1 go build

ENTRYPOINT ["/bin/sh", "-c", "/go/src/github.com/asiainfoLDP/datafoundry_oauth2/start.sh"]


