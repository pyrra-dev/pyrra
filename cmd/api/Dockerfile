# Build UI
FROM node:16 as uibuilder

WORKDIR /workspace

COPY ui/public ui/public
COPY ui/src/ ui/src
COPY ui/package.json ui
COPY ui/package-lock.json ui
COPY Makefile Makefile

RUN make ui/node_modules ui/build

# Build the kubernetes binary
FROM golang:1.16 as builder

WORKDIR /workspace

COPY cmd/api/ cmd/api/
COPY kubernetes/ kubernetes/
COPY openapi/ openapi/
COPY slo/ slo/
COPY slo/ slo/
COPY go.mod go.mod
COPY go.sum go.sum
COPY Makefile Makefile
COPY --from=uibuilder /workspace/ui/build /workspace/cmd/api/ui/build
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download
RUN make api

FROM alpine:3.14
WORKDIR /
COPY --from=builder /workspace/bin/api /usr/bin/api

ENTRYPOINT ["/usr/bin/api"]
