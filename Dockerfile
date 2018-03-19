FROM golang:onbuild
RUN mkdir -p /go/src/app
WORKDIR /go/src/app

ENV AWS_SECRET_ACCESS_KEY aws-secret-access-key
ENV AWS_ACCESS_KEY_ID aws-access-key-id
ENV AWS_REGION aws-region

RUN go get github.com/scottjbarr/sqsmv

ENTRYPOINT ["./entrypoint.sh"]

ADD entrypoint.sh .
RUN chmod 755 entrypoint.sh
RUN chmod +x entrypoint.sh