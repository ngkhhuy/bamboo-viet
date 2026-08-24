# Hướng Dẫn Đóng Góp Phát Triển (Contributing Guide)

Cảm ơn bạn đã quan tâm và muốn đóng góp cho dự án **Bamboo Viet**! Tài liệu này hướng dẫn cách thiết lập môi trường phát triển, quy chuẩn code, và quy trình gửi Pull Request.

---

## 1. Chuẩn Bị Môi Trường Phát Triển

Dự án yêu cầu:
- **Go**: Phiên bản `1.23+`
- **C/C++ Toolchain**: `gcc`, `g++`, `make`, `dpkg-dev`
- **Hệ điều hành**: Linux (Ubuntu 22.04 / 24.04, Debian 12, Fedora, Arch Linux...)

---

## 2. Quy Trình Làm Việc Với Mã Nguồn

### Bước 1: Fork và Clone
```bash
git clone https://github.com/bamboo-viet/bamboo-viet.git
cd bamboo_viet
```

### Bước 2: Tạo nhánh phát triển (Feature Branch)
```bash
git checkout -b feature/ten-tinh-nang-moi
# hoặc
git checkout -b fix/ma-so-loi
```

### Bước 3: Chạy Kiểm Thử
Trước khi viết mã nguồn mới hoặc sau khi sửa lỗi, hãy đảm bảo tất cả các bài test đều vượt qua:
```bash
make test
```
Lệnh này sẽ tự động:
1. Chạy toàn bộ Go test suite (`engine_test.go`, `engine_extended_test.go`).
2. Biên dịch thư viện Go C-Shared `libvicore.so`.
3. Biên dịch và chạy test C ABI độc lập `bin/c_test`.

### Bước 4: Đóng Gói và Thử Nghiệm Gói Cài Đặt
```bash
make deb
sudo dpkg -i bin/ibus-bamboo-viet_1.0.0_amd64.deb
ibus restart
```

---

## 3. Quy Chuẩn Đóng Góp Mã Nguồn

1. **Go Code**: Tuân thủ chuẩn `gofmt` và `go vet`. Đặt tên biến và hàm tường minh, có chú thích giải thích lý do xử lý kỹ thuật.
2. **C / C++ ABI**: Các hàm xuất ra trong `libvicore/main.go` phải đảm bảo an toàn bộ nhớ, null-terminated strings, và không làm rò rỉ bộ nhớ (memory leaks).
3. **Thêm Test Case**: Bất kỳ tính năng mới hoặc sửa lỗi nào cũng cần có unit test tương ứng trong `engine_extended_test.go` hoặc `tests/c_test.c`.

---

## 4. Gửi Pull Request (PR)

1. Commit thay đổi với thông điệp rõ ràng:
   ```bash
   git commit -m "fix(engine): resolve selection boundary issue in search bar"
   ```
2. Đẩy nhánh lên GitHub:
   ```bash
   git push origin feature/ten-tinh-nang-moi
   ```
3. Tạo Pull Request trên GitHub và mô tả chi tiết:
   - Mục đích của thay đổi.
   - Các ứng dụng/môi trường đã thử nghiệm.
   - Kết quả chạy `make test`.
