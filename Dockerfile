FROM golang:alpine
RUN apk add --update gcc musl-dev bash
RUN mkdir /3dps
COPY dockersetup.sh /3dps/setup.sh
COPY 3dps.go /3dps/3dps.go
COPY go.mod /3dps/go.mod

EXPOSE 9991
ENTRYPOINT ["bash", "/3dps/setup.sh"]
