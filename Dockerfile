FROM golang:1.12 as build

RUN mkdir /tmp/mage \
&& cd /tmp/mage \
&& wget -qO mage.tar.gz https://github.com/magefile/mage/releases/download/v1.8.0/mage_1.8.0_Linux-64bit.tar.gz \
&& tar xzf mage.tar.gz \
&& cp mage /usr/bin/

WORKDIR /app
COPY . .
RUN mage build

FROM gcr.io/distroless/base

COPY --from=build /app/crccheck /
WORKDIR /tmp

ENTRYPOINT ["/crccheck"]