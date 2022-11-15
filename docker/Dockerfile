FROM golang:alpine
RUN apk add --update gcc musl-dev curl
RUN mkdir /src
RUN mkdir /3DPS

COPY docker/setup.sh /src/setup.sh
RUN chmod +x /src/setup.sh

COPY src/*  /src/

RUN [ "sh", "/src/setup.sh" ]

EXPOSE 9991
HEALTHCHECK CMD if [[ "$(curl localhost:9991 -s | md5sum)" == "cfdf6a1ec3ca70cc1d00b7820c4ba64f  -" ]]; then; exit 0; else; exit 1; fi
ENTRYPOINT ["/3DPS/server", "-db", "/db/levels.db"]
