# --- Giai đoạn 1: BUILDER ---
# Sử dụng image golang đầy đủ để làm môi trường biên dịch.
FROM golang:1.23-alpine AS builder

# Thiết lập thư mục làm việc.
WORKDIR /app

# Cài đặt các gói cần thiết CHO VIỆC BUILD.
# Chúng sẽ không tồn tại trong image cuối cùng.
RUN apk add --no-cache git

# Tối ưu cache cho các dependency.
COPY go.mod go.sum ./
RUN go mod download

# Sao chép toàn bộ source code.
COPY . .

# Biên dịch ứng dụng.
# - CGO_ENABLED=0: Tắt CGO để tạo ra một file binary tĩnh, không phụ thuộc vào thư viện C của hệ thống.
# -o /app/main: Chỉ định file output.
# ./cmd/server: Đường dẫn đến package main của bạn.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main ./cmd/server

# --- Giai đoạn 2: FINAL ---
# Bắt đầu từ một image nền siêu nhẹ.
FROM alpine:latest

# Cài đặt ca-certificates để ứng dụng có thể thực hiện các kết nối HTTPS an toàn.
RUN apk --no-cache add ca-certificates

# Thiết lập thư mục làm việc.
WORKDIR /app

# Sao chép file binary đã được biên dịch từ giai đoạn 'builder'.
# Đây là file duy nhất từ source code của bạn được đưa vào image cuối cùng.
COPY --from=builder /app/main .

# Mở port 9001.
EXPOSE 9001

# Lệnh để chạy ứng dụng khi container khởi động.
CMD ["./main"]