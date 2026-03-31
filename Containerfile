ARG FEDORA_VERSION="43"
ARG COREDNS_VERSION=v1.12.0

#===============================================================================

FROM quay.io/fedora/fedora:${FEDORA_VERSION} AS builder

RUN dnf install -y --setopt=install_weak_deps=False --no-docs \
    ca-certificates git make

RUN { \
      GO_VERSION=$(curl -s 'https://go.dev/VERSION?m=text' | sed -ne 's/^go//p'); \
      curl -# -L https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz | \
        tar -C /usr/local -zx; \
    }
ENV PATH=/usr/local/go/bin:$PATH

ARG COREDNS_VERSION
RUN git clone --depth 1 --branch ${COREDNS_VERSION} \
    https://github.com/coredns/coredns.git /coredns

WORKDIR /coredns/

RUN go get github.com/wranders/coredns-filter

RUN sed -i '/^cache:cache/i filter:github.com/wranders/coredns-filter' plugin.cfg

RUN make

RUN useradd coredns --no-log-init -U -M -s /sbin/nologin
RUN chown coredns:coredns coredns
RUN setcap 'cap_net_bind_service=+ep' coredns
RUN mkdir user && \
    echo $(grep coredns /etc/group) > user/group && \
    echo $(grep coredns /etc/passwd) > user/passwd && \
    chown root:root user/{group,passwd} && \
    chmod 0644 user/{group,passwd}

#===============================================================================

FROM scratch

LABEL org.opencontainers.image.source="https://github.com/wranders/coredns-filter" \
    org.opencontainers.image.authors="W Anders <w@doubleu.codes>" \
    org.opencontainers.image.title="coredns-filter" \
    org.opencontainers.image.description="Sinkholing in CoreDNS" \
    org.opencontainers.image.licenses="MIT"

COPY --from=builder /coredns/coredns /

COPY --from=builder /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem \
    /etc/ssl/certs/ca-certificates.crt

COPY --from=builder /sbin/nologin /sbin/

COPY --from=builder /coredns/user/group /coredns/user/passwd /etc/

EXPOSE 53/tcp 53/udp 443/tcp 853/tcp

USER coredns

ENTRYPOINT [ "/coredns" ]
