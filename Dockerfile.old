FROM       arken/gom-base
MAINTAINER Damien Metzler <dmetzler@nuxeo.com>

RUN go get github.com/arkenio/gogeta
WORKDIR /usr/local/go/src/github.com/arkenio/gogeta
RUN git checkout master
RUN gom install
RUN gom test
RUN gom build

EXPOSE 7777
ENTRYPOINT ["/usr/local/go/src/github.com/arkenio/gogeta/gogeta", "-etcdAddress", "http://etcd:2379/", "-alsologtostderr=true"]
