FROM rust:1.31

VOLUME [ /friday ]
WORKDIR /friday


RUN apt-get update && \
    apt-get -y install cmake protobuf-compiler


ENTRYPOINT ["/usr/bin/build_casperlabs_ee_linux.sh"]
STOPSIGNAL SIGTERM

COPY build_casperlabs_ee_linux.sh /usr/bin/build_casperlabs_ee_linux.sh