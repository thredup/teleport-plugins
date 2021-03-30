FROM golang AS builder

# RUN git clone https://github.com/gravitational/teleport-plugins.git \
#     && cd teleport-plugins \
#     && make access-slack

RUN mkdir teleport-plugins
ADD ./ teleport-plugins
RUN cd teleport-plugins \
    && make access-slack

FROM debian:buster
RUN mkdir /app \
            && apt-get update \
            && apt-get install wget -y \
            && wget -P /usr/local/share/ca-certificates/cacert.org http://www.cacert.org/certs/root.crt http://www.cacert.org/certs/class3.crt \
            && update-ca-certificates \
            && groupadd -r teleport && useradd -r -g teleport teleport \
            && mkdir -p /var/lib/teleport/cfg \
            && mkdir -p /var/lib/teleport/certs \
            && mkdir -p /var/lib/teleport/tls
WORKDIR /app/
USER teleport
COPY --from=builder /go/teleport-plugins/access/slack/build/teleport-slack .
ENTRYPOINT ["./teleport-slack", "start", "--config=/var/lib/teleport/cfg/teleport-slackbot.toml", "--insecure-no-tls"]
