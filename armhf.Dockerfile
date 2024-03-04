# Stage 0 : Build the C library
FROM debian:bookworm

ENV OPENCV_VER=4.8.1-2
ENV arch=armhf

WORKDIR /foundry

RUN apt-get update -y && apt-get install -y \
    build-essential \
    cmake \
    git \
    wget \
    dpkg-dev \
    ffmpeg \
    libjpeg62-turbo-dev \
    libgtk2.0-0 \
    libgdk-pixbuf2.0-0 \
    pkg-config

RUN wget https://theheads.sfo2.digitaloceanspaces.com/shared/builds/${arch}/opencv_${OPENCV_VER}_${arch}.deb

RUN dpkg -i /foundry/opencv_${OPENCV_VER}_${arch}.deb

RUN git clone https://github.com/jgarff/rpi_ws281x.git \
  && cd rpi_ws281x \
  && mkdir build \
  && cd build \
  && cmake -D BUILD_SHARED=OFF -D BUILD_TEST=OFF .. \
  && cmake --build . \
  && make install

RUN wget https://go.dev/dl/go1.21.3.linux-armv6l.tar.gz
RUN tar -xzvf go1.21.3.linux-armv6l.tar.gz
ENV PATH="$PATH:/foundry/go/bin"

WORKDIR /build

COPY codelab codelab
COPY packager packager
COPY platform platform
COPY protobuf protobuf

WORKDIR /build/heads

COPY heads/camera ./camera
RUN (cd camera && go mod download)

COPY heads/go.mod ./go.mod
COPY heads/go.sum ./go.sum

RUN go mod download

RUN (cd camera && go build -o `mktemp -d` ./cmd/camera)