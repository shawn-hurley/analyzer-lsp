FROM quay.io/shawn_hurley/jdtls-server:testing
COPY  ./ /analyzer-lsp
COPY provider_container_settings.json /analyzer-lsp/provider_settings.json

WORKDIR /analyzer-lsp
COPY m2.xml ~/.m2/settings.xml

RUN cp java-rule-addon.core-1.0.0-SNAPSHOT.jar /jdtls/java-rule-addon/java-rule-addon.core/target/java-rule-addon.core-1.0.0-SNAPSHOT.jar

# CMD [ "go", "run", "main.go", "--error-on-violation"]
CMD [ "go", "run", "main.go", "--rules", "demo-rules/local-storage.windup-rewrite.yaml" ]
