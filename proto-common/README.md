# Proto Common

Thư mục này chứa các file proto dùng chung cho tất cả services.

## Cấu trúc

```
proto-common/
└── google/
    └── api/
        ├── annotations.proto
        └── http.proto
```

## Sử dụng

Các file trong thư mục này được import bởi tất cả services thông qua:

```protobuf
import "google/api/annotations.proto";
```

## Build

Khi generate proto files, cần thêm `--proto_path=../proto-common`:

```bash
protoc \
  --go_out=gen/go \
  --go-grpc_out=gen/go \
  --grpc-gateway_out=gen/go \
  --proto_path=proto \
  --proto_path=../proto-common \
  proto/your-service.proto
```

## Kong Gateway

Trong Kong container, thư mục này được mount vào `/etc/kong/proto/google` để tất cả services có thể import `google/api/annotations.proto`.
