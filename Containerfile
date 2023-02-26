FROM --platform=$BUILDPLATFORM registry.fedoraproject.org/fedora:37 as BUILDER
RUN dnf install -y --setopt=install_weak_deps=False --nodocs \
    ca-certificates git make
ARG BUILDARCH
RUN curl -L https://go.dev/dl/go1.20.1.linux-${BUILDARCH}.tar.gz | tar -C /usr/local -zx
ENV PATH /usr/local/go/bin:$PATH

RUN git clone https://github.com/coredns/coredns.git /coredns
WORKDIR /coredns/
RUN go get -u github.com/wranders/coredns-filter
RUN sed -i '/^cache:cache/i filter:github.com/wranders/coredns-filter' plugin.cfg
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} make

RUN useradd coredns --no-log-init -U -M -s /sbin/nologin
RUN chown coredns:coredns coredns
RUN setcap 'cap_net_bind_service=+ep' coredns
RUN mkdir user && \
    echo $(grep coredns /etc/group) > user/group && \
    echo $(grep coredns /etc/passwd) > user/passwd && \
    chown root:root user/{group,passwd} && \
    chmod 0644 user/{group,passwd}

FROM --platform=$TARGETPLATFORM scratch
LABEL org.opencontainers.image.source="https://github.com/wranders/coredns-filter" \
    org.opencontainers.image.authors="W Anders <w@doubleu.codes>" \
    org.opencontainers.image.title="coredns-filter" \
    org.opencontainers.image.description="Sinkholing in CoreDNS" \
    org.opencontainers.image.licenses="MIT"
COPY --from=BUILDER /coredns/coredns /
COPY --from=BUILDER /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem \
    /etc/ssl/certs/ca-certificates.crt
COPY --from=BUILDER /sbin/nologin /sbin/
COPY --from=BUILDER /coredns/user/group /coredns/user/passwd /etc/
EXPOSE 53/udp 443 853
USER coredns
ENTRYPOINT [ "/coredns" ]
