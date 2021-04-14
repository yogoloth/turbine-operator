FROM caninjas/hystrix_exporter:v0.1
COPY ./config.yml /config.yml
ENTRYPOINT ["/hystrix_exporter", "-c", "/config.yml"]
