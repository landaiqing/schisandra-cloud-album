# to build this docker image:
#   docker build --build-arg OPENCV_VERSION="4.11.0" -f aisvc.Dockerfile -t schisandra-ai-server .
#   docker build --build-arg OPENCV_VERSION="4.x" --build-arg OPENCV_FILE="https://github.com/opencv/opencv/archive/refs/heads/4.x.zip" --build-arg OPENCV_CONTRIB_FILE="https://github.com/opencv/opencv_contrib/archive/refs/heads/4.x.zip" -f opencv.Dockerfile -t schisandra-cloud-album-server .

FROM golang:1.23.5-bullseye AS builder

LABEL maintainer="landaiqing <<landaiqing@126.com>>"

ENV TZ=Asia/Shanghai \
    DEBIAN_FRONTEND=noninteractive

RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone && \
    sed -i 's|http://deb.debian.org/debian|https://mirrors.tuna.tsinghua.edu.cn/debian|g' /etc/apt/sources.list && \
    apt-get update && apt-get install -y --no-install-recommends \
        tzdata git build-essential cmake pkg-config unzip curl ca-certificates \
        libcurl4-openssl-dev libssl-dev libturbojpeg-dev \
        libpng-dev libtiff-dev nasm libblas-dev libatlas-base-dev\
        libdlib-dev libjpeg62-turbo-dev liblapack-dev && \
        rm -rf /var/lib/apt/lists/*

ARG OPENCV_VERSION="4.11.0"

ENV OPENCV_VERSION=$OPENCV_VERSION

ARG OPENCV_FILE="https://github.com/opencv/opencv/archive/${OPENCV_VERSION}.zip"

ENV OPENCV_FILE=$OPENCV_FILE

ARG OPENCV_CONTRIB_FILE="https://github.com/opencv/opencv_contrib/archive/${OPENCV_VERSION}.zip"

ENV OPENCV_CONTRIB_FILE=$OPENCV_CONTRIB_FILE

RUN curl -Lo opencv.zip ${OPENCV_FILE} && \
      unzip -q opencv.zip && \
      curl -Lo opencv_contrib.zip ${OPENCV_CONTRIB_FILE} && \
      unzip -q opencv_contrib.zip && \
      rm opencv.zip opencv_contrib.zip && \
      cd opencv-${OPENCV_VERSION} && \
      mkdir build && cd build && \
      cmake -D CMAKE_BUILD_TYPE=RELEASE \
      -D WITH_IPP=OFF \
      -D WITH_OPENGL=OFF \
      -D WITH_QT=OFF \
      -D WITH_FREETYPE=ON \
      -D CMAKE_INSTALL_PREFIX=/usr/local \
      -D OPENCV_EXTRA_MODULES_PATH=../../opencv_contrib-${OPENCV_VERSION}/modules \
      -D OPENCV_ENABLE_NONFREE=ON \
      -D WITH_JASPER=OFF \
      -D WITH_TBB=ON \
      -D BUILD_JPEG=ON \
      -D WITH_SIMD=ON \
      -D ENABLE_LIBJPEG_TURBO_SIMD=ON \
      -D BUILD_DOCS=OFF \
      -D BUILD_EXAMPLES=OFF \
      -D BUILD_TESTS=OFF \
      -D BUILD_PERF_TESTS=ON \
      -D BUILD_opencv_java=NO \
      -D BUILD_opencv_python=NO \
      -D BUILD_opencv_python2=NO \
      -D BUILD_opencv_python3=NO \
      -D OPENCV_GENERATE_PKGCONFIG=ON .. && \
      make -j $(nproc --all) && \
      make preinstall && make install && ldconfig && \
      cd / && rm -rf opencv*

WORKDIR /app

COPY . .

ENV CGO_ENABLED=1 \
    CGO_CFLAGS="-I/usr/local/include/opencv4" \
    CGO_CPPFLAGS="-I/usr/local/include" \
    CGO_LDFLAGS="-L/usr/local/lib -lopencv_core -lopencv_face -lopencv_videoio -lopencv_imgproc -lopencv_highgui -lopencv_imgcodecs -lopencv_objdetect -lopencv_features2d -lopencv_video -lopencv_dnn -lopencv_xfeatures2d" \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct

RUN go mod download && \
    go build -ldflags="-w -s" -o schisandra-ai-server ./app/aisvc/rpc/aisvc.go

FROM debian:bullseye-slim AS runtime

ENV TZ=Asia/Shanghai \
    DEBIAN_FRONTEND=noninteractive

RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone && \
    apt-get update && apt-get install -y --no-install-recommends \
    tzdata libjpeg62-turbo libblas3 liblapack3 libdlib-dev  libtiff5 && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /usr/local/lib /usr/local/lib/

COPY --from=builder /usr/local/include/opencv4 /usr/local/include/opencv4/

COPY --from=builder /app/schisandra-ai-server .

COPY --from=builder /app/app/aisvc/rpc/etc ./rpc/etc

COPY --from=builder /app/app/aisvc/resources ./resources

ENV LD_LIBRARY_PATH=/usr/local/lib

EXPOSE 8888

CMD ["./schisandra-ai-server"]
