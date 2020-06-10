FROM frolvlad/alpine-glibc:alpine-3.7_glibc-2.26
MAINTAINER Greg Szabo <greg@tendermint.com>

RUN apk update && \
    apk upgrade && \
    apk --no-cache add curl jq file libstdc++

VOLUME [ /nodef ]
WORKDIR /nodef
EXPOSE 26656 26657 6060
ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["start"]
STOPSIGNAL SIGTERM

COPY wrapper.sh /usr/bin/wrapper.sh

