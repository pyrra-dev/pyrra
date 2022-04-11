# Build UI
FROM node:17.8.0 as uibuilder

WORKDIR /workspace

COPY ui/public ui/public
COPY ui/src/ ui/src
COPY ui/package.json ui
COPY ui/package-lock.json ui
COPY ui/tsconfig.json ui
COPY Makefile Makefile

RUN make ui/node_modules ui/build

# Build the kubernetes binary
FROM golang:1.17.8 as builder

WORKDIR /workspace

COPY kubernetes/ kubernetes/
COPY openapi/ openapi/
COPY slo/ slo/
COPY *.go ./
COPY go.mod ./
COPY go.sum ./
COPY Makefile Makefile
COPY --from=uibuilder /workspace/ui/build /workspace/ui/build
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download
RUN make pyrra

FROM alpine:3.15.4
WORKDIR /
COPY --from=builder /workspace/pyrra /usr/bin/pyrra

ENTRYPOINT ["/usr/bin/pyrra"]
