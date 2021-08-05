# Build UI
FROM node:16 as uibuilder

WORKDIR /workspace
COPY ui ui
COPY Makefile Makefile
RUN make ui/node_modules ui/build

# Build the kubernetes binary
FROM golang:1.16 as builder

WORKDIR /workspace
COPY . .
COPY --from=uibuilder /workspace/ui/build /workspace/cmd/api/ui/build
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Build
RUN make kubernetes

# Use distroless as minimal base image to package the kubernetes binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/bin/kubernetes .
USER nonroot:nonroot

ENTRYPOINT ["/kubernetes"]
