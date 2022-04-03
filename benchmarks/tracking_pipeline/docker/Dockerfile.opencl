FROM ubuntu:20.04 AS builder
ARG DEBIAN_FRONTEND=noninteractive
ARG TARGETPLATFORM
ARG OPENCV_VERSION

RUN apt-get update && apt-get -y install wget python3.8 python3.8-dev python3-distutils && ln -s -f /usr/bin/python3.8 /usr/bin/python3
RUN wget -qO /tmp/get-pip.py https://bootstrap.pypa.io/get-pip.py && python3 /tmp/get-pip.py
RUN python3 -m pip install numpy

# Install Build Tools and Libraries
RUN apt-get -y install make cmake build-essential wget cmake pkg-config libarchive-tools \
    libjpeg-dev libpng-dev libtiff-dev \
    libavcodec-dev libavformat-dev libavutil-dev libswscale-dev libavresample-dev \
    libv4l-dev libxvidcore-dev libx264-dev libatlas-base-dev gfortran
    
# Download and unpack opencv + opencv_contrib sources
RUN wget -qO opencv.zip https://github.com/opencv/opencv/archive/${OPENCV_VERSION}.zip \
    && wget -qO opencv_contrib.zip https://github.com/opencv/opencv_contrib/archive/${OPENCV_VERSION}.zip \
    && mkdir opencv opencv_contrib \
    && bsdtar --strip-components=1 -C opencv/ -xvf opencv.zip \
    && bsdtar --strip-components=1 -C opencv_contrib/ -xvf opencv_contrib.zip

# Change into workdir for cmake build of opencv
WORKDIR /opencv/build/

# Configure cmake build
# TODO:
RUN mkdir install
RUN case $TARGETPLATFORM in \
    linux/amd64) \
        cmake -D CMAKE_BUILD_TYPE=RELEASE -D CMAKE_INSTALL_PREFIX=install -D OPENCV_EXTRA_MODULES_PATH=../../opencv_contrib/modules -D OPENCV_ENABLE_NONFREE=ON -D PYTHON_EXECUTABLE=/usr/bin/python3 -D WITH_GTK=OFF -D BUILD_JAVA=OFF -D BUILD_TESTS=OFF -D BUILD_PERF_TESTS=OFF -D CPU_DISPATCH=AVX,AVX2 -D WITH_OPENCL=ON -D OPENCV_DNN_OPENCL=ON .. \
    ;; \
    linux/arm64) \
        cmake -D CMAKE_BUILD_TYPE=RELEASE -D CMAKE_INSTALL_PREFIX=install -D OPENCV_EXTRA_MODULES_PATH=../../opencv_contrib/modules -D OPENCV_ENABLE_NONFREE=ON -D PYTHON_EXECUTABLE=/usr/bin/python3 -D WITH_GTK=OFF -D BUILD_JAVA=OFF -D BUILD_TESTS=OFF -D BUILD_PERF_TESTS=OFF -D ENABLE_NEON=ON -D WITH_OPENCL=ON -D OPENCV_DNN_OPENCL=ON .. \
    ;; \
    esac

# Build and install opencv
RUN make -j$(nproc) && make install

FROM ubuntu:20.04
LABEL maintainer="sbaeurle, simon.baeurle@tum.de"
ARG INTEL_GMMLIB=https://github.com/intel/compute-runtime/releases/download/21.42.21270/intel-gmmlib_21.2.1_amd64.deb
ARG INTEL_IGC_CORE=https://github.com/intel/intel-graphics-compiler/releases/download/igc-1.0.8744/intel-igc-core_1.0.8744_amd64.deb
ARG INTEL_IGC_OPENCL=https://github.com/intel/intel-graphics-compiler/releases/download/igc-1.0.8744/intel-igc-opencl_1.0.8744_amd64.deb
ARG INTEL_OPENCL_ICD=https://github.com/intel/compute-runtime/releases/download/21.42.21270/intel-opencl-icd_21.42.021270_amd64.deb
ARG AMD_DRIVER_URL=https://drivers.amd.com/drivers/linux/amdgpu-pro-20.20-1098277-ubuntu-20.04.tar.xz
ARG NVIDIA_DRIVER=https://us.download.nvidia.com/XFree86/Linux-x86_64/470.63.01/NVIDIA-Linux-x86_64-470.63.01.run
ARG DEBIAN_FRONTEND=noninteractive

# Install required dependencies
RUN apt-get update && apt-get -y install ffmpeg libavcodec-extra

# Move compiled opencv to image and relink libraries
COPY --from=builder /opencv/build/install/lib /lib

RUN apt-get -y install wget kmod apt-utils build-essential python3.8 python3-distutils && ln -s -f /usr/bin/python3.8 /usr/bin/python3
RUN wget -qO /tmp/get-pip.py https://bootstrap.pypa.io/get-pip.py && python3 /tmp/get-pip.py
RUN python3 -m pip install numpy

# Install Intel Driver
WORKDIR /tmp/intel
RUN wget -q ${INTEL_GMMLIB} ${INTEL_IGC_CORE} ${INTEL_IGC_OPENCL} ${INTEL_OPENCL_ICD}
RUN dpkg -i /tmp/intel/*.deb

# Install AMD Radeon Driver / Potentially switch to ROCm (TODO)
WORKDIR /tmp/amd
RUN wget --referer=http://support.amd.com -qO /tmp/amd/amd-driver.tar.xz ${AMD_DRIVER_URL} && tar -Jxvf /tmp/amd/amd-driver.tar.xz
RUN dpkg --add-architecture i386
RUN cd amdgpu-pro-* && ./amdgpu-install -y -opencl=legacy,rocr --headless --no-dkms
RUN mkdir -p /etc/OpenCL/vendors && \
    echo "libamdocl64.so" > /etc/OpenCL/vendors/amdocl64.icd

# Prepare for NVIDIA Container Runtime
RUN mkdir -p /etc/OpenCL/vendors && \
    echo "libnvidia-opencl.so.1" > /etc/OpenCL/vendors/nvidia.icd
ENV NVIDIA_VISIBLE_DEVICES all
ENV NVIDIA_DRIVER_CAPABILITIES compute,utility

# Remove Mesa OpenCL Implementation
RUN rm -f /etc/OpenCL/vendors/mesa.icd && rm -rf /tmp/*

# Temporary fix to included python3 dev headers to build grpcio
RUN apt-get -y install python3.8-dev

RUN apt-get clean

WORKDIR /

RUN ldconfig