
# GoFrameX - Framework Web Hiện Đại Cho Go

**GoFrameX** là một framework web mạnh mẽ, có tính mô-đun và khả năng mở rộng cao dành cho các ứng dụng Go. Dự án tuân theo nguyên tắc kiến trúc sạch (*clean architecture*) và cung cấp các công cụ thiết yếu để xây dựng các ứng dụng cấp doanh nghiệp.

## 💡 Tính Năng Nổi Bật

- 🏗️ **Kiến trúc sạch**: Xây dựng theo nguyên lý thiết kế hướng miền (*domain-driven design*) và kiến trúc sạch
- 🔌 **Thiết kế mô-đun**: Dễ dàng mở rộng và tùy chỉnh với hệ thống component có thể cắm rời (*pluggable*)
- 🚀 **Hiệu năng cao**: Tối ưu hóa với các thực hành hiện đại trong Go
- 🔒 **Ưu tiên bảo mật**: Tích hợp các tính năng và middleware bảo mật sẵn có
- 🔄 **Tích hợp cơ sở dữ liệu**: Hỗ trợ GORM cho thao tác cơ sở dữ liệu
- 📦 **Hỗ trợ cache**: Tích hợp Redis cho caching
- 📝 **Logging có cấu trúc**: Logging nâng cao với Logrus
- ⚙️ **Quản lý cấu hình**: Linh hoạt với Viper
- 🌐 **HTTP Client tích hợp**: Dựa trên thư viện Resty
- ⏰ **Lên lịch tác vụ**: Hỗ trợ Cron Job cho các tác vụ định kỳ

## 🗂️ Cấu Trúc Dự Án

\`\`\`plaintext
├── cmd/                    # Điểm khởi động ứng dụng
│   └── Main.go            # Tệp main của ứng dụng
├── internal/              # Mã nguồn chính (private)
│   ├── application/       # Các quy tắc nghiệp vụ ứng dụng
│   ├── bootstrap/         # Khởi tạo ứng dụng
│   ├── domain/            # Các quy tắc nghiệp vụ cốt lõi (domain)
│   ├── infrastructure/    # Giao tiếp với hệ thống bên ngoài (DB, cache, ...)
│   └── interfaces/        # Các cơ chế giao tiếp (HTTP, gRPC, ...)
└── pkg/                   # Thư viện dùng chung (có thể dùng lại)
    ├── cache/             # Tiện ích cache
    ├── config/            # Tiện ích cấu hình
    ├── database/          # Tiện ích cơ sở dữ liệu
    ├── http/              # Tiện ích HTTP
    ├── logger/            # Tiện ích logging
    └── scheduler/         # Tiện ích lên lịch (cron)
\`\`\`

## 🔧 Phụ Thuộc

- Go 1.23.1 hoặc mới hơn
- Gin Web Framework
- GORM - ORM cho Go
- Redis
- PostgreSQL
- Logrus
- Viper
- Resty
- Cron

## 🚀 Bắt Đầu

1. **Clone dự án:**

\`\`\`bash
git clone https://github.com/xhkzeroone/goframex.git
\`\`\`

2. **Cài đặt các gói phụ thuộc:**

\`\`\`bash
go mod download
\`\`\`

3. **Cấu hình ứng dụng:**

- Sao chép file \`resources/config.yml\` và điều chỉnh các thiết lập nếu cần

4. **Chạy ứng dụng:**

\`\`\`bash
go run cmd/Main.go
\`\`\`

## ⚙️ Cấu Hình

Ứng dụng sử dụng file \`config.yml\` trong thư mục \`resources\` để quản lý cấu hình.

## 🔌 Các Mô-đun

### HTTP Server (\`pkg/http/ginx\`)
- Xây dựng dựa trên Gin framework
- Middleware cấu hình được
- Xử lý Request/Response tiện lợi

### Database (\`pkg/database/gormx\`)
- Tích hợp GORM để thao tác cơ sở dữ liệu
- Hỗ trợ connection pool
- Hỗ trợ migration

### Cache (\`pkg/cache/redisx\`)
- Tích hợp Redis
- Hỗ trợ cấu hình chiến lược caching

### Logger (\`pkg/logger/logrusx\`)
- Logging có cấu trúc
- Hỗ trợ nhiều mức độ log
- Format JSON
- Ẩn dữ liệu nhạy cảm

### Scheduler (\`pkg/scheduler/cronx\`)
- Lên lịch tác vụ định kỳ bằng cron
- Quản lý các tác vụ chạy nền

## 🤝 Đóng Góp

Mọi đóng góp đều được hoan nghênh! Hãy gửi Pull Request nếu bạn muốn đóng góp vào dự án.

## 📜 Giấy Phép

Dự án này được cấp phép theo giấy phép MIT — xem file \`LICENSE\` để biết thêm chi tiết.

## 👤 Tác Giả

**xhkzeroone**
