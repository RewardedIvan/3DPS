FROM golang:alpine
RUN apk add --update gcc musl-dev
RUN mkdir /3dps
COPY setup.sh /3dps/setup.sh
COPY 3dps.go /3dps/3dps.go

EXPOSE 9991
ENTRYPOINT ["bash", "/3dps/setup.sh"]