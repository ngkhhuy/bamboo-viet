# Checklist Kiểm Thử Trước Phát Hành (Release Checklist)

Tài liệu này dùng để kiểm tra chất lượng (QA) toàn diện trước khi phát hành phiên bản mới của bộ gõ **Bamboo Viet**.

---

## 1. Automated Testing & Build Validation
- [ ] Chạy `make test` vượt qua 100% không có lỗi.
- [ ] Chạy `make test-c` kiểm tra liên kết C ABI với `libvicore.so` thành công.
- [ ] Chạy `make deb` xuất ra gói `bin/ibus-bamboo-viet_<version>_amd64.deb` hợp lệ.
- [ ] Kiểm tra metadata trong file `.deb`: `dpkg-deb -I <file.deb>`.
- [ ] Kiểm tra cấu trúc file giải nén: `dpkg-deb -c <file.deb>`.

---

## 2. Core Functional Testing (IBus & Wayland)
- [ ] **Kiểu gõ Telex:** Gõ chuẩn các từ phức tạp: `tiếng`, `việt`, `đường`, `nghiêng`, `khuyến`, `quốc`, `trường`, `phương`.
- [ ] **Kiểu gõ VNI:** Gõ chuẩn `tie6ng1` $\rightarrow$ `tiếng`, `vie6t5` $\rightarrow$ `việt`.
- [ ] **Thay đổi dấu tốc độ cao:** Gõ `toan` $\rightarrow$ `toán` $\rightarrow$ `toàn` $\rightarrow$ `toản` $\rightarrow$ `toãn` $\rightarrow$ `toạn` không bị kẹt phím.
- [ ] **Bảng mã:** Unicode dựng sẵn, Unicode tổ hợp.
- [ ] **Lọc phím tắt hệ thống:** `Ctrl+C`, `Ctrl+V`, `Ctrl+A`, `Alt+Tab`, phím `Super` không bị nuốt hoặc chèn ký tự rác.

---

## 3. Per-App Real-World Verification
- [ ] **Chat Apps (Zalo, Slack, Telegram, Messenger Web):**
  - Gõ một từ tiếng Việt (ví dụ `tiếng việt`) rồi nhấn phím `Enter`.
  - Xác nhận **KHÔNG bị lặp lại từ cuối cùng** (`việt việt`).
- [ ] **Trình duyệt (Google Chrome / Brave / Edge):**
  - Gõ từ khóa vào thanh tìm kiếm hoặc thanh địa chỉ URL.
  - Xác nhận khi có dropdown gợi ý (autocomplete), con trỏ không bị nhảy hoặc nuốt chữ.
- [ ] **Code Editors (VSCode / Cursor):**
  - Gõ chú thích hoặc biến tiếng Việt trong file code.
  - Xác nhận IntelliSense popup không làm lệch vị trí con trỏ.
- [ ] **Văn phòng (LibreOffice Writer / Google Docs):**
  - Soạn thảo văn bản dài với tốc độ gõ nhanh.
  - Xác nhận không có hiện tượng giật con trỏ hoặc nuốt chữ.
- [ ] **Terminal (Kitty, Alacritty, GNOME Terminal):**
  - Gõ lệnh tiếng Anh và tiếng Việt trong terminal không bị đơ hoặc nhảy dòng.

---

## 4. Configuration & Packaging Validation
- [ ] `bamboo-viet-config status` hiển thị thông tin chính xác.
- [ ] `bamboo-viet-config set-method VNI` chuyển đổi thành công.
- [ ] `bamboo-viet-config set-mode SurroundingText` lưu cấu hình đúng.
- [ ] `bamboo-viet-config restart` khởi động lại IBus trơn tru.
- [ ] Thao tác cài đặt `sudo dpkg -i` và gỡ bỏ `sudo apt purge` hoạt động sạch sẽ.
