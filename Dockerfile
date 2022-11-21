FROM golang:1.18 as builder
WORKDIR /analyzer-lsp
# TODO limit to prevent unnecessary rebuilds
COPY  . /analyzer-lsp
RUN make build

# The unofficial base image w/ jdtls and gopls installed
FROM quay.io/shawn_hurley/jdtls-server 

WORKDIR /analyzer-lsp
# TODO limit to prevent unnecessary rebuilds
COPY  . /analyzer-lsp
COPY --from=builder /analyzer-lsp/konveyor-analyzer /analyzer-lsp/konveyor-analyzer
COPY provider_container_settings.json /analyzer-lsp/provider_settings.json


RUN go install golang.org/x/tools/gopls@latest
RUN [ "/analyzer-lsp/konveyor-analyzer", "--rules", "demo-rules/local-storage.windup-rewrite.yaml" ]

# Uncomment to enable delve debugging
# RUN go install github.com/go-delve/delve/cmd/dlv@latest
# RUN rm rule-example.yaml
# CMD [ "dlv", "debug", "main.go", "--", "--rules", "demo-rules/local-storage.windup-rewrite.yaml" ]
