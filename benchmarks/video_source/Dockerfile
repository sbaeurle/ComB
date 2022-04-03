FROM ubuntu:20.04 as builder
ENV DEBIAN_FRONTEND=noninteractive

# Install dev dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev libgstreamer-plugins-bad1.0-dev libgstrtspserver-1.0-dev \
    gstreamer1.0-plugins-base gstreamer1.0-plugins-good gstreamer1.0-plugins-bad gstreamer1.0-plugins-ugly \
    gstreamer1.0-libav gstreamer1.0-x gstreamer1.0-alsa gstreamer1.0-gl

# Setup meson and ninja
RUN apt-get install -y python3 python3-pip && pip install ninja meson

COPY server.c meson.build ./

RUN meson setup build && cd build && meson compile

FROM ubuntu:20.04
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    gstreamer1.0-plugins-base gstreamer1.0-plugins-good gstreamer1.0-plugins-bad gstreamer1.0-plugins-ugly \
    gstreamer1.0-libav gstreamer1.0-x gstreamer1.0-alsa gstreamer1.0-gl libgstrtspserver-1.0

COPY --from=builder build/server server

CMD ["./server", "data/cam1_dataset.mp4"]