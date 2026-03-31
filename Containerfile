ARG FEDORA_VERSION="43"
ARG COREDNS_VERSION="1.14.2"

#===============================================================================

FROM quay.io/fedora/fedora:${FEDORA_VERSION} AS builder

RUN dnf install -y --setopt=install_weak_deps=False --no-docs \
    make \
    util-linux

ARG TARGETARCH
RUN { \
      GO_VERSION=$(curl -s 'https://go.dev/VERSION?m=text' | sed -ne 's/^go//p'); \
      GO_ARCH=$TARGETARCH; \
      if [[ "$TARGETARCH" == "arm" ]]; then GO_ARCH="arm64"; fi; \
      curl -# -L https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz | \
        tar -C /usr/local -zx; \
    }
ENV PATH=/usr/local/go/bin:$PATH

ARG COREDNS_VERSION
RUN mkdir /coredns && \
    curl -# -L https://github.com/coredns/coredns/archive/refs/tags/v${COREDNS_VERSION}.tar.gz \
    | tar -C /coredns -zx --strip-components=1

WORKDIR /coredns/

RUN go get github.com/wranders/coredns-filter

RUN sed -i '/^cache:cache/i filter:github.com/wranders/coredns-filter' plugin.cfg

RUN make

RUN mkdir -p /scratch/etc/ && \
    touch /scratch/etc/{passwd,group} && \
    useradd coredns \
      --prefix=/scratch \
      --no-log-init \
      --system \
      --user-group \
      --no-create-home \
      --shell=/sbin/nologin

RUN setcap 'cap_net_bind_service=+ep' coredns

#===============================================================================

FROM scratch

LABEL org.opencontainers.image.source="https://github.com/wranders/coredns-filter"
LABEL org.opencontainers.image.authors="W Anders <w@doubleu.codes>"
LABEL org.opencontainers.image.title="coredns-filter"
LABEL org.opencontainers.image.description="Sinkholing in CoreDNS"
LABEL org.opencontainers.image.licenses="MIT"

COPY --from=builder /scratch /

COPY --from=builder /coredns/coredns /

COPY --from=builder /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem \
    /etc/ssl/certs/ca-certificates.crt

COPY --from=builder /sbin/nologin /sbin/

EXPOSE 53/tcp 53/udp 443/tcp 853/tcp

USER coredns

ENTRYPOINT [ "/coredns" ]
