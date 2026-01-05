# Quick Start: Running Agent for Development

## Vấn Đề Đã Fix ✅

~~Khi chạy `go run main.go`, agent báo lỗi:~~
```
Failed to load configuration: invalid configuration: org_id is required
```

**Đã fix:** Environment variables giờ sẽ override các giá trị rỗng trong config file.

## Giải Pháp Nhanh

### Option 1: Set Environment Variables ✅ **RECOMMENDED**

```bash
# Set credentials
export ORG_ID="test-org"
export INSTALL_TOKEN="test-token-123"

# Run agent
cd cmd/agent
go run main.go
```

**Lưu ý:** Ngay cả khi config file đã tồn tại với `org_id` = `""`, env vars sẽ override nó.

### Option 2: Edit Config File

```bash
# Agent đã tạo config file tại:
# Windows: C:\ProgramData\unitechio\Agent\config.json
# Linux: /etc/your-agent/config.json

# Edit file và thay đổi:
{
  "org_id": "test-org",           # <- Thay từ ""
  "install_token": "test-token",  # <- Thay từ ""
  ...
}

# Sau đó run lại
go run main.go
```

### Option 3: Xóa Config File Cũ

```bash
# Xóa config file để agent tạo mới từ env vars
rm "C:\ProgramData\unitechio\Agent\config.json"

# Set env vars và run
export ORG_ID="test-org"
export INSTALL_TOKEN="test-token"
go run main.go
```

## Lưu Ý

⚠️ **Agent sẽ fail khi bootstrap** vì credentials không hợp lệ:
```
Bootstrap failed: 401 Unauthorized
```

Đây là **bình thường** khi test local. Agent đã start thành công, chỉ là không connect được backend.

## Để Test Đầy Đủ

Cần credentials thực từ backend:
1. Đăng ký organization trên admin portal
2. Lấy `org_id` và `install_token`
3. Set vào env vars hoặc config file
4. Run agent

## Tham Khảo

Xem hướng dẫn chi tiết tại: [CONFIGURATION_DEPLOYMENT.md](./CONFIGURATION_DEPLOYMENT.md)
