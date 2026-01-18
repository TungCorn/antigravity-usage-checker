---
description: Tạo release mới với changelog và test đầy đủ
---

# Workflow Release

Thực hiện các bước sau để tạo release mới:

## 1. Xác nhận phiên bản mới
- Hỏi user phiên bản mới là gì (ví dụ: v2.1.0)
- Kiểm tra phiên bản mới nhất trong CHANGELOG.md để tránh trùng

## 2. Cập nhật CHANGELOG.md
- Thêm section mới cho phiên bản ở đầu file (sau dòng header)
- Format theo chuẩn đã có:
```markdown
## [vX.X.X] Tên release

### Added
- Tính năng mới

### Changed
- Thay đổi

### Fixed
- Sửa lỗi

---
```
- Hỏi user mô tả các thay đổi nếu chưa biết

## 3. Build thử
// turbo
- Chạy lệnh: `go build -o agusage.exe ./cmd/agusage/`
- Đảm bảo build thành công
- Xóa file build thử sau khi xong

## 4. Commit changes
- Stage tất cả thay đổi: `git add .`
- Commit với message: `chore: prepare release vX.X.X`

## 5. Tạo tag và push
// turbo
- Tạo tag: `git tag vX.X.X`
// turbo
- Push code: `git push origin main`
// turbo
- Push tag: `git push origin vX.X.X`

## 6. Xác nhận hoàn thành
- Thông báo user workflow đã hoàn thành
- Cung cấp link đến GitHub Actions để theo dõi build: https://github.com/tungcorn/antigravity-usage-checker/actions
