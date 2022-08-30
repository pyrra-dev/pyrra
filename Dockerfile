FROM --platform="${BUILDPLATFORM:-linux/amd64}" docker.io/busybox:1.35.0 as builder

ARG TARGETOS TARGETARCH TARGETVARIANT

WORKDIR /app
COPY dist dist

# NOTICE: See goreleaser.yml for the build paths.
RUN env && if [ "${TARGETARCH}" = 'amd64' ]; then \
        cp "dist/pyrra_${TARGETOS}_${TARGETARCH}_${TARGETVARIANT:-v1}/pyrra" . ; \
    elif [ "${TARGETARCH}" = 'arm' ]; then \
        cp "dist/pyrra_${TARGETOS}_${TARGETARCH}_${TARGETVARIANT##v}/pyrra" . ; \
    else \
        cp "dist/pyrra_${TARGETOS}_${TARGETARCH}/pyrra" . ; \
    fi
RUN chmod +x pyrra

FROM --platform="${TARGETPLATFORM:-linux/amd64}"  docker.io/alpine:3.16.2 AS runner
WORKDIR /
COPY --chown=0:0 --from=builder /app/pyrra /usr/bin/pyrra
USER nobody

ENTRYPOINT ["/usr/bin/pyrra"]
