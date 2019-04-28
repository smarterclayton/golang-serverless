FROM registry.access.redhat.com/ubi7/ubi-minimal:latest

ENV VERSION=1.12.1 \
    GOCACHE=/go/path/.cache \
    GOPATH=/go/path \
    GOROOT=/usr/local/go
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin
WORKDIR /go/src/
COPY run-go /usr/bin/
COPY serve /go/src/serve/
RUN microdnf install -y tar gzip && microdnf clean all && rm -rf /var/cache/yum/* && \
    curl https://storage.googleapis.com/golang/go$VERSION.linux-amd64.tar.gz | tar -C /usr/local -xzf - && \
    rm -rf $GOROOT/{pkg/linux_amd64_race,test,doc,api}/* \
           $GOROOT/pkg/tool/linux_amd64/{vet,doc,cover,trace,nm,fix,test2json,objdump} \
           $GOROOT/bin/godoc \
           /var/lib/rpm/Packages && \
    find /usr/share/locale/ -name tar.mo | xargs rm && \
    mkdir -p $GOPATH/bin && \
    chmod g+xw -R /go && \
    chmod g+xw -R $(go env GOROOT)
ENTRYPOINT ["/usr/bin/run-go"]