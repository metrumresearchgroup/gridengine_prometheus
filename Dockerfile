FROM golang:alpine AS compiler
RUN    apk add --no-cache git \
    && git clone https://github.com/metrumresearchgroup/gridengine_prometheus.git \
    && cd gridengine_prometheus/ \
    && go build -o gridengine_exporter cmd/server/main.go

FROM debian:buster-slim
RUN    apt-get -qq update >/dev/null 2>&1 \
    && apt-get -qq install musl cpp libmunge2 libssl1.1 libjemalloc2 libbsd0 libevent-2.1-6 libgnutls-dane0 liblockfile-bin liblockfile1 libunbound8 netbase >/dev/null 2>&1 \
    && apt-get -qq clean >/dev/null 2>&1 \
    && apt-get -qq autoclean >/dev/null 2>&1 \
    && ln -s ld-musl-x86_64.so.1 /lib/libc.musl-x86_64.so.1 \
    && rm -rf /var/lib/apt/lists/* \
    && echo "#!/usr/bin/env bash" > /usr/local/bin/launch_sge_exporter.sh \
    && echo '/usr/local/bin/gridengine_exporter --sge_root /var/lib/gridengine --sge_cell $SGE_CELL --port $PORT' >> /usr/local/bin/launch_sge_exporter.sh \
    && chmod +x /usr/local/bin/launch_sge_exporter.sh
COPY --from=compiler /go/gridengine_prometheus/gridengine_exporter /usr/local/bin/
#COPY ./launch_sge_exporter.sh /usr/local/bin/
ENV SGE_CELL default
ENV PORT 9081
ENTRYPOINT /usr/local/bin/launch_sge_exporter.sh
