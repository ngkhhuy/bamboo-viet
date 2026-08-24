# 🎋 Bamboo Viet — Bộ Gõ Tiếng Việt Hiện Đại Cho Linux

[![CI](https://github.com/bamboo-viet/bamboo-viet/actions/workflows/ci.yml/badge.svg)](https://github.com/bamboo-viet/bamboo-viet/actions/workflows/ci.yml)
[![License: GPL-3.0](https://img.shields.io/badge/License-GPL%203.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](go.mod)
[![Platform: Linux](https://img.shields.io/badge/Platform-Linux%20(Wayland%20%2F%20X11)-orange)](https://ubuntu.com)

**Bamboo Viet** là bộ gõ tiếng Việt mã nguồn mở thế hệ mới cho Linux (hỗ trợ cả **IBus** và **Fcitx5**). Dự án kế thừa và phát triển từ `ibus-bamboo` & `ibus-lotus`, tập trung sửa dứt điểm các lỗi kinh niên của các bộ gõ tiếng Việt trên Linux hiện đại (Wayland, Electron, LibreOffice, các app Chat).

---

## 🚀 Các Cải Tiến & Lỗi Đã Sửa Dứt Điểm

| Lỗi / Vấn đề trên Linux | Trạng thái trước đây | Bamboo Viet |
|---|---|---|
| **Lặp từ khi nhấn Enter trong Chat** (Messenger, Slack, Telegram, Zalo) | ❌ Lặp từ cuối | ✅ **Đã sửa** (Cơ chế Commit-before-hide) |
| **Lỗi giật / Popup trên Wayland native** | ❌ Bị lỗi popup tương tác | ✅ **Đã sửa** (Tách luồng & tối ưu Fcitx5) |
| **Nhảy con trỏ, nuốt chữ trong LibreOffice / Docs** | ❌ Trễ nhịp D-Bus (200ms) | ✅ **Đã sửa** (Tăng tốc độ phản hồi gấp 6 lần) |
| **Lỗi gợi ý / Autocomplete trong ô tìm kiếm & URL** | ❌ Nuốt chữ khi có gợi ý | ✅ **Đã sửa** (Selection-aware & Token boundary) |
| **Electron & Chromium Hardening** (Chrome, VSCode, Zalo PC) | ⚠️ Dễ lệch dấu, mất chữ | ✅ **Đã sửa** (Built-in App Preset Profiles) |
| **Xung đột phím tắt hệ thống** (`Ctrl+C`, `Alt+Tab`, `Super`) | ❌ Bị nhận nhầm ký tự gõ | ✅ **Đã sửa** (Bộ lọc `isValidState` mở rộng) |

---

## 📦 Cài Đặt Nhanh (1 Lệnh)

### Cách 1: Cài đặt từ gói `.deb` (Ubuntu, Debian, Linux Mint, Pop!_OS)

1. Tải file cài đặt `.deb` mới nhất từ mục **Releases** hoặc tự build:
```bash
sudo dpkg -i ibus-bamboo-viet_1.0.0_amd64.deb
```
2. Khởi động lại IBus:
```bash
ibus restart
```

### Cách 2: Biên dịch và cài đặt từ mã nguồn

```bash
# Clone repository
git clone https://github.com/bamboo-viet/bamboo-viet.git
cd bamboo_viet

# Chạy test và biên dịch
make test
make deb

# Cài đặt
sudo dpkg -i bin/ibus-bamboo-viet_1.0.0_amd64.deb
ibus restart
```

---

## 🖥️ Giao Diện Bảng Điều Khiển (GUI UniKey Style)

Bamboo Viet trang bị sẵn **Bảng điều khiển đồ họa (GUI) chuẩn phong cách UniKey trên Windows**:
- Mở từ Menu ứng dụng: Tìm kiếm **"Bamboo Viet Control Panel"**
- Hoặc chạy từ terminal:
```bash
bamboo-viet-gui
```
- **Các tính năng trên giao diện:**
  - **Khung chính:** Chọn nhanh Bảng mã (Unicode, TCVN3, VNI Windows...), Kiểu gõ (Telex, VNI, Simple Telex), Chế độ gõ (Surrounding Text, Pre-edit...).
  - **Nút [ Mở rộng >> ] / [ << Thu gọn ]:** Bật/tắt tùy chọn nâng cao (Kiểm tra chính tả, Đặt dấu kiểu mới, Tự động khôi phục từ sai, Sửa lỗi chat Enter...).
  - **Nút [ Bảng gõ tắt... ]:** Trình soạn thảo danh sách từ gõ tắt (Macro) trực quan.

---

## 🛠️ Công Cụ Cấu Hình Dòng Lệnh (`bamboo-viet-config`)

Bộ gõ đi kèm công cụ CLI `bamboo-viet-config` giúp quản lý cấu hình tiện lợi qua terminal:

```bash
# Xem trạng thái cấu hình hiện tại
bamboo-viet-config status

# Đổi kiểu gõ sang Telex hoặc VNI
bamboo-viet-config set-method Telex
bamboo-viet-config set-method VNI

# Đổi chế độ gõ mặc định
bamboo-viet-config set-mode SurroundingText
bamboo-viet-config set-mode Preedit

# Xem danh sách cấu hình tự động cho các ứng dụng
bamboo-viet-config list-presets

# Gán cấu hình riêng cho ứng dụng tùy chọn
bamboo-viet-config set-app "google-chrome:Google-chrome" 2

# Khởi động lại IBus
bamboo-viet-config restart
```

---

## ⌨️ Bảng Phím Tắt & Chế Độ Gõ

* **Bật / Tắt bộ gõ:** `<Super>` + `<Space>` hoặc `<Ctrl>` + `<Space>` (tùy cấu hình desktop của bạn).
* **Chuyển chế độ gõ nhanh cho cửa sổ hiện tại:** `<Shift>` + `~`
* **Các chế độ gõ hỗ trợ:**
  1. `PreeditIM`: Chế độ gạch chân gõ chữ an toàn.
  2. `SurroundingTextIM`: Chế độ xóa lùi thông minh (Tối ưu cho Chrome, VSCode, LibreOffice, Zalo).
  3. `ForwardAsCommitIM`: Chuyển tiếp phím trực tiếp (Tối ưu cho Terminal: Kitty, Alacritty, Gnome-terminal).
  4. `UsIM`: Tắt hoàn toàn bộ gõ cho cửa sổ chỉ định (Chế độ tiếng Anh).

---

## 🧩 Cấu Trúc Dự Án

```text
bamboo_viet/
├── bin/                    # Binaries và gói cài đặt (.deb, libvicore.so, ibus-engine)
├── cmd/config_tool/        # Công cụ CLI bamboo-viet-config
├── config/                 # Quản lý cấu hình & Built-in App Presets
├── docs/                   # Tài liệu ma trận lỗi & checklist
├── fcitx5/                 # Fcitx5 C++ Addon (hỗ trợ Wayland text-input-v3)
├── libvicore/              # Thư viện Go C-Shared xuất chuẩn C ABI (libvicore.so)
├── packaging/              # Cấu trúc đóng gói Debian packaging
├── tests/                  # Bộ test kiểm thử C ABI & unit tests
├── engine.go               # IBus Engine core implementation
├── engine_backspace.go     # Backspace handling & text diff
├── engine_extended_test.go # Test suite mở rộng
└── Makefile                # Build automation (make test, make deb, make vicore)
```

---

## 🤝 Đóng Góp Phát Triển

Chúng tôi rất hoan nghênh mọi đóng góp từ cộng đồng! Vui lòng đọc [CONTRIBUTING.md](CONTRIBUTING.md) để biết thêm chi tiết về quy trình gửi Pull Request và chạy kiểm thử.

---

## ☕ Ủng Hộ Phát Triển (Donate)

Bamboo Viet là dự án phần mềm mã nguồn mở hoàn toàn miễn phí. Nếu dự án giúp ích cho công việc hàng ngày của bạn trên Linux, bạn có thể ủng hộ tác giả một ly cà phê qua Ko-fi để tiếp thêm động lực duy trì và phát triển:

[![Support me on Ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/ngkhhuy)

* **☕ Ko-fi:** [ko-fi.com/ngkhhuy](https://ko-fi.com/ngkhhuy)
* Mọi sự ủng hộ tùy tâm từ cộng đồng đều là nguồn khích lệ vô cùng to lớn đối với sự phát triển của dự án!

---

## 📄 Bản Quyền & Giấy Phép

Dự án được phát hành dưới giấy phép mã nguồn mở **GNU General Public License v3.0 (GPL-3.0)**. Xem chi tiết tại file [LICENSE](LICENSE).
Thuật toán biến đổi âm tiết tiếng Việt sử dụng `bamboo-core` (Giấy phép MIT).

