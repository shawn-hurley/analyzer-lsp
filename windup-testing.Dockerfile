FROM quay.io/shawn_hurley/jdtls-server 
COPY  ./ /analyzer-lsp
COPY provider_container_settings.json /analyzer-lsp/provider_settings.json

WORKDIR /analyzer-lsp

CMD [ "go", "run", "main.go", "--error-on-violation"]
