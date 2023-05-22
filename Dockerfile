FROM --platform="${BUILDPLATFORM:-linux/amd64}" docker.io/busybox:1.36.1 as builder

ARG TARGETOS TARGETARCH TARGETVARIANT

WORKDIR /app
COPY dist dist

# NOTICE: See goreleaser.yml for the build paths.
RUN if [ "${TARGETARCH}" = 'amd64' ]; then \
        cp "dist/pyrra_${TARGETOS}_${TARGETARCH}_${TARGETVARIANT:-v1}/pyrra" . ; \
    elif [ "${TARGETARCH}" = 'arm' ]; then \
        cp "dist/pyrra_${TARGETOS}_${TARGETARCH}_${TARGETVARIANT##v}/pyrra" . ; \
    else \
        cp "dist/pyrra_${TARGETOS}_${TARGETARCH}/pyrra" . ; \
    fi
RUN chmod +x pyrra

FROM --platform="${TARGETPLATFORM:-linux/amd64}"  docker.io/alpine:3.18.0 AS runner
WORKDIR /
COPY --chown=0:0 --from=builder /app/pyrra /usr/bin/pyrra
USER 65533

ENTRYPOINT ["/usr/bin/pyrra"]
