FROM golang:1.14-buster as builder
RUN apt-get update -q -y \
    && apt-get install -q -y --no-install-recommends \
      libusb-1.0-0-dev \
      && rm -rf /var/lib/apt/lists/*
COPY . /usb-reset
WORKDIR /usb-reset
RUN go build .

FROM debian:buster-slim
RUN apt-get update -q -y \
    && apt-get install -q -y --no-install-recommends \
      libusb-1.0-0 \
      && rm -rf /var/lib/apt/lists/*
WORKDIR /
COPY --from=builder /usb-reset/usb-reset /
CMD [ "/usb-reset" ]
