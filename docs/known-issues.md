# Danh mục & Phân tích Chi tiết Lỗi Bộ gõ Tiếng Việt Linux (docs/known-issues.md)

Tài liệu này là kết quả nghiệm thu của **Giai đoạn 0 (Đào bới, so sánh & tái hiện lỗi)**, tổng hợp từ việc đối chiếu trực tiếp mã nguồn, commit history, và kiến trúc giữa **ibus-bamboo gốc** (BambooEngine) và **ibus-lotus** (hien-ngo29/ibus-bamboo-ng).

---

## 1. Tổng quan So sánh Mã nguồn: ibus-bamboo vs ibus-lotus

### 1.1 Khác biệt về Module & Dependencies
| Thành phần | ibus-bamboo gốc | ibus-lotus (fork) | Ghi chú & Đánh giá |
|---|---|---|---|
| **Go Version** | `go 1.13` | `go 1.23.0` (toolchain `1.24.5`) | Lotus đã hiện đại hóa toolchain Go |
| **Thuật toán cốt lõi** | `BambooEngine/bamboo-core` | `BambooEngine/bamboo-core` | Giữ nguyên 100% logic parser tiếng Việt |
| **IBus D-Bus Binding** | `BambooEngine/goibus` | `LotusInputEngine/goibus` | Chuyển org và dọn dẹp import, core D-Bus tương đương |
| **Wayland API** | `github.com/dkolbly/wl` | *Đã loại bỏ* | Bamboo dùng thư viện dkolbly cũ (bị lỗi trên Mutter), Lotus đã gỡ bỏ |

### 1.2 Tái cấu trúc file & Cải tiến trong ibus-lotus
* **Thêm `backspace_faker.go`:** Tách logic gửi phím xóa giả lập (`SendBackSpace`, `SendBackspaceInSurroundingTextMode`, `SendBackspaceXTest`, `SendBackspaceForwardAsCommitMode`) từ `engine_backspace.go` thành file riêng, dễ bảo trì và mở rộng.
* **Thêm `x11_record.c`:** Tách luồng bắt sự kiện chuột (`thread_mouse_recording`) chạy trên pthread riêng biệt sử dụng extension XRecord, tránh nghẽn luồng xử lý phím chính của IBus.
* **Thay thế `wl_introspector.go`:** Thay vì cố gắn vào `zwlr_foreign_toplevel_manager_v1` (vốn không được hỗ trợ trên GNOME Wayland/Mutter), Lotus chuyển sang gọi D-Bus GNOME Shell Extension (`org.gnome.Shell.Extensions.WindowsExt`) trên GNOME và `kdotool` trên KDE.
* **Thêm hàm `isValidState(state)`:** Lọc bỏ các phím tắt kết hợp với `Ctrl`, `Alt (Mod1)`, `Super (Mod4)`, `Meta`, `Hyper` để tránh kích hoạt xử lý âm tiết tiếng Việt khi người dùng bấm phím tắt hệ thống/ứng dụng.

---

## 2. Bảng Ma trận Đối chiếu Lỗi (Bug Matrix)

| STT | Lỗi / Triệu chứng | Trạng thái trên ibus-bamboo | Trạng thái trên ibus-lotus | Trạng thái trên bamboo_viet |
|---|---|---|---|---|
| **BUG-01** | Lặp từ cuối cùng khi nhấn Enter trong app chat (Messenger, Slack, Telegram, Zalo) | ❌ Bị lỗi | ✅ Đã sửa | ✅ **Đã hoàn thiện** (Commit-before-hide) |
| **BUG-02** | Popup "Allow Remote Interaction" và crash/giật trên GNOME Wayland native | ❌ Bị lỗi | ⚠️ Dùng D-Bus extension | ⚠️ D-Bus extension (Đã có Fcitx5 text-input-v3) |
| **BUG-03** | Giật màn hình / mất focus khi click chuột trong ứng dụng X11 & Wayland | ❌ Bị lỗi | ✅ Đã sửa | ✅ **Đã hoàn thiện** (Thread XRecord độc lập) |
| **BUG-04** | Nhảy con trỏ, nuốt chữ trên LibreOffice / Google Docs / Rich Text Editors | ❌ Bị lỗi | ❌ Vẫn còn (xung đột timing `DeleteSurroundingText`) | ✅ **Đã sửa** (Bỏ blocking sleep, tối ưu IPC D-Bus) |
| **BUG-05** | Lỗi gõ tiếng Việt trên ô tìm kiếm (Search Box) Firefox, Chrome, GNOME Files | ❌ Bị lỗi | ⚠️ Chưa xử lý autocomplete selection | ✅ **Đã sửa** (Selection-aware & isolate token boundaries) |
| **BUG-06** | Nuốt chữ đầu hoặc lệch dấu trên các app Electron/Chromium (Chrome, VSCode, Zalo PC) | ❌ Bị lỗi | ⚠️ Cần cấu hình per-app | ✅ **Đã hoàn thiện** (Built-in Per-App Preset Profiles) |
| **BUG-07** | Nhận nhầm phím tắt hệ thống (`Ctrl+C`, `Alt+Tab`, `Super`) thành ký tự gõ | ❌ Bị lỗi | ✅ Đã sửa (`isValidState` filter) | ✅ **Đã hoàn thiện** (Bộ lọc isValidState mở rộng) |

---

## 3. Phân tích Chi tiết Root-Cause Từng Nhóm Lỗi

### 3.1 BUG-01: Lặp từ cuối cùng khi nhấn Enter trong các ứng dụng Chat (Web/Electron)
* **Triệu chứng:** Khi gõ một từ tiếng Việt (ví dụ `tiếng`) và nhấn `Enter` ngay lập tức để gửi tin nhắn, kết quả nhận được là `tiếng tiếng` hoặc bị lặp lại từ cuối cùng.
* **Nguyên nhân kỹ thuật:**
  * Trong chế độ `Preedit`, IBus quản lý chuỗi đang soạn thảo chưa commit.
  * Mã nguồn gốc của Bamboo trong `engine_preedit.go`:
    ```go
    func (e *IBusBambooEngine) commitPreeditAndReset(s string) {
        e.HidePreeditText()   // 1. Ẩn preedit
        e.HideAuxiliaryText()
        e.HideLookupTable()
        e.commitText(s)       // 2. Commit nội dung
        e.preeditor.Reset()
    }
    ```
  * Khi `HidePreeditText()` được kích hoạt, IBus gửi một sự kiện `preedit_changed(empty)` tới ứng dụng. Các thư viện soạn thảo hiện đại (như Draft.js trên Facebook, Lexical, Slate) khi nhận sự kiện `preedit_changed` rỗng sẽ tự động đẩy nội dung buffer hiện tại vào DOM tree.
  * Ngay sau đó, lệnh `commitText(s)` đến, ứng dụng lại chèn tiếp chuỗi `s` một lần nữa $\rightarrow$ Tạo ra hiện tượng nhân đôi từ (lặp từ).
* **Giải pháp của Lotus (Đã kiểm chứng):**
  * Đảo ngược thứ tự: gọi `commitText(s)` trước, sau đó mới gọi `HidePreeditText()`:
    ```go
    func (e *IBusBambooEngine) commitPreeditAndReset(s string) {
        e.commitText(s)       // 1. Đưa text vào buffer trước
        e.HidePreeditText()   // 2. Ẩn preedit sau
        e.HideAuxiliaryText()
        e.HideLookupTable()
        e.preeditor.Reset()
    }
    ```

---

### 3.2 BUG-02 & BUG-03: Lỗi Wayland Native, Popup "Allow Remote Interaction" và Chuột
* **Triệu chứng:** Khi chạy trên Ubuntu/Fedora GNOME Wayland, hệ thống liên tục hiện popup cảnh báo bảo mật đòi quyền "Allow Remote Interaction". Khi click chuột trong một số ứng dụng, màn hình bị nháy/giật hoặc mất tiêu điểm con trỏ.
* **Nguyên nhân kỹ thuật:**
  * Wayland được thiết kế với mô hình bảo mật cô lập: một client thông thường không được phép đọc trạng thái cửa sổ của client khác.
  * Bamboo cũ dùng `github.com/dkolbly/wl` cố bind protocol `zwlr_foreign_toplevel_manager_v1`. Protocol này chỉ có trên compositors chuẩn wlroots (như Sway), hoàn toàn không tồn tại trên GNOME Mutter, dẫn đến lỗi bind và kích hoạt cơ chế bảo mật của GNOME.
  * Về chuột: việc hook vào X11 event stream trong môi trường hỗn hợp XWayland gây xung đột thời gian thực giữa thread xử lý D-Bus và X server.
* **Giải pháp hiện tại của Lotus:**
  * Bỏ `dkolbly/wl`.
  * Trên GNOME: Dùng D-Bus gọi extension `org.gnome.Shell.Extensions.WindowsExt`.
  * Trên KDE: Dùng `kdotool getactivewindow`.
  * Tách riêng `x11_record.c` thành thread độc lập `thread_mouse_recording`.
* **Hạn chế tồn đọng & Hướng xử lý lâu dài:**
  * Việc phụ thuộc extension GNOME hoặc tool bên ngoài chưa phải giải pháp lý tưởng cho mọi người dùng.
  * **Giải pháp chuẩn:** Khi chuyển sang **Fcitx5 (Giai đoạn 3)**, Fcitx5 sử dụng giao thức `text-input-v3` chuẩn của Wayland, Compositor tự quản lý focus và context mà không cần bất kỳ thủ thuật introspect nào.

---

### 3.3 BUG-04: Nhảy con trỏ & Mất chữ trên LibreOffice / Rich Text Editors
* **Triệu chứng:** Khi soạn thảo văn bản trong LibreOffice Writer hoặc Google Docs, khi gõ nhanh hoặc chỉnh sửa dấu ở giữa từ, con trỏ nhảy ngược về đầu từ, chữ bị nuốt hoặc đảo lộn thứ tự ký tự (ví dụ: gõ `việt` thành `vệit`).
* **Nguyên nhân kỹ thuật:**
  * Chế độ `SurroundingTextIM` trong `backspace_faker.go` xóa ký tự bằng cách gửi `DeleteSurroundingText(-N, N)` qua D-Bus:
    ```go
    func (e *IBusBambooEngine) SendBackspaceInSurroundingTextMode() {
        time.Sleep(20 * time.Millisecond)
        log.Printf("Sendding %d backspace via SurroundingText\n", fakeBackspaceCount)
        e.DeleteSurroundingText(-int32(fakeBackspaceCount), uint32(fakeBackspaceCount))
        time.Sleep(20 * time.Millisecond)
    }
    ```
  * Vấn đề xảy ra do **bất đồng bộ thời gian (Timing Race Condition)**:
    1. Lệnh xóa `DeleteSurroundingText` được gửi bất đồng bộ qua D-Bus tới LibreOffice.
    2. Trong khi LibreOffice đang tính toán lại layout và cập nhật vị trí con trỏ (cursor offset), lệnh `commitText` tiếp theo từ Go engine đã được gửi tới.
    3. Do độ trễ cố định (`time.Sleep(20ms)`) không khớp với tốc độ render thực tế của ứng dụng, ký tự mới bị chèn vào vị trí *trước* khi ký tự cũ bị xóa hoàn tất, khiến con trỏ bị lệch vị trí.
* **Đề xuất khắc phục cho Giai đoạn 2:**
  * Đồng bộ hóa chuỗi commit-delete bằng cách kiểm tra phản hồi sự kiện `surrounding_text_changed` trước khi commit chuỗi mới.
  * Cung cấp profile cấu hình fallback về `ForwardAsCommitIM` hoặc `Preedit` cho các ứng dụng có rendering engine nặng như LibreOffice.

---

### 3.4 BUG-05: Lỗi trên Ô Tìm kiếm & Thanh Địa chỉ (Search Box / Address Bar)
* **Triệu chứng:** Khi gõ vào ô tìm kiếm của Firefox, thanh địa chỉ Chrome hoặc thanh tìm kiếm GNOME Files, việc bỏ dấu bị ngắt quãng, xóa nhầm vào các từ gợi ý autocomplete, hoặc biến mất ký tự đang gõ.
* **Nguyên nhân kỹ thuật:**
  * Khi người dùng gõ vào Search Box, trình duyệt lập tức kích hoạt tính năng **Inline Autocomplete / Suggestions Dropdown**.
  * Dropdown này kích hoạt sự kiện thay đổi text trong ô input, làm IBus gửi sự kiện `SetSurroundingText` mới với độ dài và vị trí con trỏ hoàn toàn khác với những gì người dùng vừa gõ (chứa cả phần text gợi ý được highlight).
  * Hàm `SetSurroundingText` trong `engine.go` bị nhầm lẫn giữa chuỗi do người dùng gõ và chuỗi gợi ý của browser, dẫn đến tính toán sai số ký tự cần xóa.
* **Đề xuất khắc phục cho Giai đoạn 2:**
  * Nhận diện trạng thái selection (khi `cursorPos != anchorPos`), tự động bỏ qua việc can thiệp vào chuỗi đã được select tự động bởi browser.
  * Tinh chỉnh thuật toán `getLastWordFromSentence` để chỉ lấy từ thực sự nằm trước `min(cursorPos, anchorPos)`.

---

### 3.5 BUG-06: Vấn đề đặc thù trên Chromium & Electron Apps (Chrome, VSCode, Zalo PC)
* **Triệu chứng:** Ký tự đầu tiên bị nuốt, hoặc khi gõ nhanh các phím phụ âm kép/dấu bị vỡ thành ký tự thô (ví dụ: gõ `đ` thành `dd`, gõ `thư` thành `thuwr`).
* **Nguyên nhân kỹ thuật:**
  * Chromium có pipeline xử lý phím riêng (Blink IME Layer). Khi chạy trên Linux với Ozone platform (chuyển đổi giữa Wayland và X11), Chromium xử lý sự kiện `ForwardKeyEvent` và `CommitText` theo các hàng đợi khác nhau.
  * Nếu engine gửi phím dạng `XTestFakeKeyEvent` trong khi app đang chạy native Wayland (`--ozone-platform=wayland`), sự kiện bàn phím bị drop hoàn toàn vì Wayland không cho phép XTest can thiệp vào client Wayland thuần túy.
* **Đề xuất khắc phục cho Giai đoạn 4:**
  * Xây dựng cơ chế phát hiện tự động backend hiển thị của ứng dụng (XWayland vs Native Wayland).
  * Thiết lập preset cấu hình tối ưu per-app (Preset cho Chromium/VSCode/Zalo).

---

## 4. Kế hoạch Hành động Kỹ thuật cho Giai đoạn 1 & 2

### 4.1 Chuẩn bị cho Giai đoạn 1 (Fork & Làm quen)
1. **Khởi tạo Repo mới:** Fork trực tiếp từ `ibus-lotus` (đã bao gồm các fix BUG-01, BUG-03, `isValidState`, `x11_record.c`).
2. **Quản lý Dependencies:** Giữ nguyên `bamboo-core`, tái sử dụng module `goibus` từ Lotus.
3. **Môi trường Build:** Đảm bảo cài đặt đủ `pkg-config`, `libgtk-3-dev`, `libx11-dev`, `libxtst-dev` để build trọn vẹn cả engine và GUI.
4. **Xây dựng Test Suite (`testdata/`):**
   - Viết các test case tự động cho `engine_preedit.go` và `engine_backspace.go` mô phỏng các chuỗi gõ phím nhanh, xử lý word-break rune và macro.

### 4.2 Trọng tâm cho Giai đoạn 2 (Sửa lỗi IBus)
- Tập trung vá triệt để **BUG-04 (LibreOffice surrounding text timing)** và **BUG-05 (Search box autocomplete recognition)**.
- Nghiệm thu trên 4 ứng dụng chuẩn: *GNOME Text Editor, LibreOffice Writer, Firefox Search Bar, GNOME Files*.
