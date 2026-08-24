# Kế hoạch dự án: Bộ gõ tiếng Việt cho Linux (Fcitx5 + IBus)

**Bối cảnh:** Việt Nam đang siết bản quyền phần mềm mạnh từ 05/2026, tạo làn sóng doanh nghiệp chuyển sang Linux. Các bộ gõ tiếng Việt hiện có (ibus-bamboo, ibus-unikey) đều đang bị bỏ ngỏ hoặc ngừng cập nhật, gây ra hàng loạt lỗi ảnh hưởng trực tiếp đến năng suất người dùng văn phòng. Đây là khoảng trống thật, đúng thời điểm.

**Người thực hiện:** Solo, bán thời gian đều đặn, nền tảng C/C++, sẵn sàng dùng Go cho phần lõi.

---

## 1. Mục tiêu & phạm vi

**Mục tiêu Giai đoạn đầu (MVP):** một bộ gõ tiếng Việt chạy ổn định trên Fcitx5 + Wayland, không còn các lỗi kinh điển của bamboo (mất chữ, nhảy con trỏ, lỗi search box, lỗi trên Electron apps), có GUI cấu hình cơ bản, và **được chính bạn dùng làm driver hàng ngày**.

**Ngoài phạm vi ở giai đoạn đầu:** hỗ trợ đa nền tảng (macOS/Windows), tối ưu cho mọi ứng dụng, mô hình kinh doanh — những cái này quyết định *sau* khi có dữ liệu dùng thật.

---

## 2. Quyết định kỹ thuật cốt lõi

### 2.1 Đính chính quan trọng: bamboo-core thực ra là Go, không phải C

Kiểm tra trực tiếp source code và tài liệu HACKING.adoc của ibus-bamboo xác nhận: dự án viết **đa phần bằng Go**, chỉ 3 file xử lý X11 (`x11_clipboard.c`, `x11_introspector.c`, `x11_keyboard.c`) là C. Hai thư viện Go chính:

- **`BambooEngine/goibus`** — tương tác với IBus qua D-Bus (đã có sẵn, không cần tự viết engine IBus từ đầu bằng `godbus/dbus` như dự tính ban đầu)
- **`BambooEngine/bamboo-core`** — xử lý thuật toán tiếng Việt, tác giả xác nhận "code này đã ổn định, không cần động tới"
- `dkolbly/wl` — Wayland API bọc bằng Go, nhưng không còn được cập nhật, có thể cần fork/duy trì riêng

→ Điều này đổi hẳn kết luận: **không cần cgo, không cần binding** cho phần core + IBus, vì cả hai đã là Go sẵn — chỉ việc fork và dùng trực tiếp như một Go module.

### 2.2 Ngôn ngữ theo từng thành phần

| Thành phần | Ngôn ngữ | Lý do |
|---|---|---|
| Core engine (thuật toán tiếng Việt) | Go — fork trực tiếp `bamboo-core` | Đã ổn định qua nhiều năm, không viết lại |
| Engine IBus | Go — fork trực tiếp `goibus` + code engine của ibus-bamboo/ibus-lotus | Đã có sẵn, hoạt động qua D-Bus |
| Addon Fcitx5 | C++ mỏng, gọi vào Go qua `-buildmode=c-shared` | Xác nhận trực tiếp từ tác giả bamboo: Fcitx5 addon bắt buộc qua Fcitx5 C++ API hoặc tự bọc C++ API cho ngôn ngữ khác — không có đường vòng nào khác |

Build `c-shared` cho phép Go xuất hàm ra C ABI (`//export`), C++ shim chỉ việc gọi các hàm đó — không cần viết lại logic tiếng Việt bằng C++.

### 2.3 Không viết thuật toán tiếng Việt từ đầu — và không nên viết lại toàn bộ dự án

Thuật toán đặt dấu, gõ tắt, xử lý edge case đã được bamboo-core mài giũa nhiều năm qua vô số bug report thực tế. Chính tác giả gốc của ibus-bamboo cũng cảnh báo điều này trong tài liệu HACKING.adoc, dẫn lại bài viết kinh điển "Things You Should Never Do" của Joel Spolsky về việc không nên viết lại toàn bộ mã nguồn từ đầu.

**Quan trọng — đã có người tiếp nối dự án:** fork tên **ibus-lotus** (github.com/hien-ngo29/ibus-bamboo-ng) đang chủ động fix bug từ giữa 2025, đã sửa được lỗi lặp từ cuối cùng ở một số app chat, lỗi bắt sự kiện chuột gây phiền trên ứng dụng Wayland native, lỗi giật màn hình khi click chuột — và đã chạy được trên Wayland. Nên **fork từ ibus-lotus thay vì ibus-bamboo gốc đã ngừng phát triển**, vừa kế thừa các fix đã có, vừa có thể phối hợp với người đang tích cực làm thay vì làm trùng lặp.

### 2.4 Giấy phép

ibus-bamboo (và có khả năng cao bamboo-core, goibus cũng vậy — cần kiểm tra license riêng từng repo trước khi bắt đầu) là GPL 3.0. Nếu fork và dùng trực tiếp, toàn bộ dự án cần tương thích GPL — hợp lý nếu định hướng dự án là mã nguồn mở, đúng tinh thần một công cụ hạ tầng dùng chung.

---

## 3. Kiến trúc tổng thể

```
                    ┌───────────────────────────────┐
                    │  bamboo-core (Go, fork trực tiếp) │
                    │  - Telex/VNI parser              │
                    │  - Đặt dấu, gõ tắt (đã ổn định)   │
                    └───────────────┬───────────────────┘
                                    │  import trực tiếp (Go module), KHÔNG cgo
                ┌───────────────────┴─────────────────────┐
                │                                          │
      ┌─────────▼──────────────┐            ┌──────────────▼───────────────┐
      │  ibus-lotus engine (Go) │            │  libvicore.so                 │
      │  dùng goibus, qua D-Bus │            │  build với -buildmode=c-shared │
      │  → kế thừa fix có sẵn   │            │  xuất hàm C ABI                │
      └──────────────────────────┘           └──────────────┬────────────────┘
                                                              │
                                                 ┌────────────▼─────────────┐
                                                 │  Fcitx5 addon (C++ mỏng)   │
                                                 │  gọi vào libvicore.so      │
                                                 └────────────────────────────┘
```

Core logic (bamboo-core) và engine IBus (goibus) đều đã có sẵn bằng Go — việc của bạn là fork, sửa/bổ sung, không viết lại từ đầu. Chỉ lớp Fcitx5 cần code mới, và chỉ ở mức shim mỏng.

---

## 4. Lộ trình theo giai đoạn

| Giai đoạn | Nội dung | Thời gian ước tính* | Tiêu chí hoàn thành |
|---|---|---|---|
| **0. Đào bới, so sánh & tái hiện lỗi** | Build cả ibus-bamboo gốc và ibus-lotus, đối chiếu bug nào lotus đã fix, tái hiện các lỗi còn lại (Chrome, search box...) để xác định root cause | 2–3 tuần | Có danh sách lỗi còn tồn đọng kèm nguyên nhân kỹ thuật rõ ràng, không đoán mò |
| **1. Fork & làm quen codebase** | Fork ibus-lotus, đọc kỹ `bamboo-core`, `goibus`, viết test case cho các bug đã biết ở Giai đoạn 0 | 3–4 tuần | Build được từ source theo HACKING.adoc, engine chạy ổn định trên máy bạn |
| **2. Sửa lỗi còn tồn đọng trên IBus** | Vá các bug engine IBus chưa được ibus-lotus xử lý, ưu tiên nhóm liên quan surrounding-text/pre-edit | 3–4 tuần | Gõ ổn định trên GNOME Text Editor, LibreOffice, không còn mất chữ/nhảy con trỏ |
| **3. Fcitx5 addon (C++ mỏng)** | `-buildmode=c-shared` xuất bamboo-core ra C ABI, viết addon C++ theo Fcitx5 API, ưu tiên `text-input-v3` cho Wayland | 3–5 tuần | Gõ ổn định trên KDE/Wayland, không cần pre-edit gạch chân ở hầu hết app |
| **4. Electron apps hardening** | Test & sửa riêng cho Chrome, VSCode, Zalo PC, 1 app chat web (Slack/Teams) | 3–4 tuần | 5 app trên gõ mượt, không mất chữ, không lệch vị trí |
| **5. GUI cấu hình + đóng gói** | Config UI cơ bản (kiểu gõ, gõ tắt, bảng mã), gói `.deb` + Flatpak | 3–4 tuần | Cài đặt bằng 1 lệnh, không cần sửa file text tay |
| **6. Dogfood mở rộng** | Tự dùng hàng ngày, mời vài người quen dùng thử, cân nhắc phối hợp với maintainer ibus-lotus | Liên tục | Đủ dữ liệu để quyết định có mở rộng thành sản phẩm hay không |

*Ước tính cho bán thời gian đều đặn (~10–15h/tuần). Vì tái sử dụng codebase Go có sẵn thay vì viết engine từ đầu, tổng thời gian tới MVP dùng được hàng ngày rút ngắn còn khoảng 3–4 tháng.

---

## 5. Cấu trúc thư mục đề xuất

```
vi-linux-ime/
├── core/                  # Go package — logic tiếng Việt
│   ├── engine.go
│   ├── bamboo_cgo.go      # cgo bindings vào bamboo-core
│   └── engine_test.go
├── ibus/                  # Go — IBus engine qua D-Bus
│   └── main.go
├── fcitx5/                # C++ shim mỏng
│   ├── addon.cpp
│   ├── addon.h
│   └── CMakeLists.txt
├── third_party/
│   └── bamboo-core/       # submodule, giữ nguyên license GPL
├── testdata/              # bộ test case tái hiện các bug đã biết
├── packaging/
│   ├── debian/
│   └── flatpak/
└── docs/
    └── known-issues.md    # log lỗi từ Giai đoạn 0, đối chiếu khi fix
```

---

## 6. Rủi ro & cách giảm thiểu

| Rủi ro | Giảm thiểu |
|---|---|
| Học godbus/dbus + cgo cùng lúc mất thời gian hơn dự kiến | Dành trọn Giai đoạn 0 để làm quen, không vội sang code chính |
| Lặp lại bug cũ vì không hiểu rõ root cause | Kỷ luật ở Giai đoạn 0 — không sang giai đoạn sau khi chưa root-cause xong |
| Nản vì làm một mình, không thấy tiến độ | Dogfood sớm (ngay cuối Giai đoạn 2) — tự trải nghiệm cải thiện là động lực thật |
| Mất phương hướng giữa "dự án học tập" và "sản phẩm" | Chưa quyết định quy mô — đúng, cứ để MVP tự nói lên câu trả lời ở Giai đoạn 6 |

---

## 7. Tài nguyên tham khảo

- ibus-bamboo (gốc, đã ngừng phát triển): https://github.com/BambooEngine/ibus-bamboo
- **ibus-lotus (fork đang tiếp tục phát triển, nên bắt đầu từ đây):** https://github.com/hien-ngo29/ibus-bamboo-ng
- HACKING.adoc (tài liệu kiến trúc chính thức, đọc kỹ trước khi code): https://github.com/BambooEngine/ibus-bamboo/blob/master/docs/HACKING.adoc
- Thảo luận về tình trạng dự án: https://github.com/BambooEngine/ibus-bamboo/issues/590
- BambooEngine/goibus (thư viện Go tương tác IBus qua D-Bus)
- BambooEngine/bamboo-core (thư viện Go xử lý thuật toán tiếng Việt)
- fcitx5-unikey (baseline đối chiếu cho phần Fcitx5): tìm trên Fcitx5 addon repo chính thức
- Fcitx5 addon development docs: fcitx.github.io
- Wayland `text-input-v3` protocol spec
- Dependencies build theo HACKING.adoc: `make`, `go`, `gtk3`, `libX11`, `libXtst` (build) — `dbus`, `ibus` (runtime)
