FROM golang:1.5
MAINTAINER      Damien METZLER <dmetzler@nuxeo.com>

RUN go get github.com/tools/godep
RUN go get github.com/mjibson/esc
RUN CGO_ENABLED=0 go install -a std
ENV APP_DIR $GOPATH/src/github.com/arkenio/gogeta

# Set the entrypoint as the binary, so `docker run <image>` will behave as the binary
EXPOSE 7777
ENTRYPOINT ["/gogeta", "-etcdAddress", "http://etcd:2379/", "-alsologtostderr=true"]

ADD . $APP_DIR
# Compile the binary and statically link
RUN cd $APP_DIR && \
    esc -o statictemplate.go -prefix templates templates && \
    CGO_ENABLED=0 godep restore && \
    godep go build -o /gogeta -ldflags '-w -s'
