# Installation Instructions

1. Install protoc from https://pypi.org/project/grpcio-tools/
2. Run `protoc --proto_path=/usr/local/include/ --proto_path=. --python_out=<output-path> detection.proto tracking.proto` to compile proto definitions into python source code