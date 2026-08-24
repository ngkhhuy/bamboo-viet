# Hướng dẫn chi tiết từng giai đoạn — Bộ gõ tiếng Việt cho Linux

Tài liệu này đi sâu vào cách thực hiện từng giai đoạn trong `ke-hoach-bo-go-tieng-viet-linux.md`. Đọc file đó trước để nắm bối cảnh và kiến trúc tổng thể.

**Nguyên tắc xuyên suốt:** fork và sửa, không viết lại. Điểm khởi đầu là [ibus-lotus](https://github.com/hien-ngo29/ibus-bamboo-ng) (fork đang tiếp tục phát triển của ibus-bamboo), không phải ibus-bamboo gốc đã ngừng cập nhật.

---

## Giai đoạn 0 — Đào bới, so sánh & tái hiện lỗi (2–3 tuần)

### Mục tiêu
Biết chính xác lỗi nào đã được ibus-lotus fix, lỗi nào còn tồn đọng, và nguyên nhân kỹ thuật của từng lỗi còn lại — trước khi viết bất kỳ dòng code nào.

### Việc cần làm

**0.1 Chuẩn bị môi trường test**
- Dựng 2 máy/VM Ubuntu: một chạy session X11, một chạy Wayland (để so sánh hành vi — README của bamboo đã cảnh báo IBus trên Wayland yếu hơn X11)
- Cài đặt các dependency build theo `docs/HACKING.adoc`: `make`, `go`, `gtk3`, `libX11`, `libXtst` (build-time); `dbus`, `ibus` (runtime)

**0.2 Build cả hai bản để đối chiếu**
```
# ibus-bamboo gốc
git clone https://github.com/BambooEngine/ibus-bamboo
cd ibus-bamboo
sudo make build PREFIX=/usr

# ibus-lotus (fork đang phát triển)
git clone https://github.com/hien-ngo29/ibus-bamboo-ng
cd ibus-bamboo-ng
# build tương tự, đối chiếu Makefile/README của repo này vì có thể khác đôi chút
```
Chạy `./ibus-bamboo-engine` (hoặc tên binary tương ứng của lotus) sau khi `ibus-daemon` đã khởi động, theo đúng hướng dẫn trong HACKING.adoc.

**0.3 Lập bảng đối chiếu bug**

Dùng danh sách lỗi đã biết làm điểm khởi đầu, test trên cả 2 bản, ghi lại trạng thái:

| Lỗi | Trên ibus-bamboo gốc | Trên ibus-lotus | Ghi chú |
|---|---|---|---|
| Lặp từ cuối cùng trong app chat khi nhấn Enter | Còn | Đã fix (theo issue #574) | Xác nhận lại trên app chat bạn thực dùng |
| "Allow Remote Interaction" phiền khi dùng app Wayland native | Còn | Đã fix (theo issue #487) | Test cụ thể trên vài app Wayland-native |
| Giật màn hình khi click chuột | Còn | Đã fix (theo issue #500) | |
| Con trỏ nhảy khi gõ trên LibreOffice/Facebook | Chưa rõ | Chưa rõ | Cần tự test |
| Không nhận diện search box | Chưa rõ | Chưa rõ | Test trên GNOME Files, Firefox address bar |
| Lỗi trên Chrome | Chưa rõ | Chưa rõ | Test riêng, xem có cần cài thêm gói như ibus-unikey từng cần không |
| Không gõ được sau update lên bản Ubuntu mới | Chưa rõ | Chưa rõ | |

Bổ sung thêm các lỗi bạn tự gặp phải trong quá trình dùng hàng ngày.

**0.4 Đọc source theo đúng thứ tự để hiểu luồng chạy**

Theo gợi ý của chính tác giả trong HACKING.adoc: bắt đầu từ `main()`, theo dõi các hàm nó gọi.

Thứ tự đọc đề xuất:
1. `main.go` — điểm vào, khởi tạo engine
2. `engine.go` — vòng lặp xử lý phím chính
3. `engine_preedit.go` — logic chế độ gạch chân (pre-edit) — đây là nơi nhiều bug tập trung
4. `engine_backspace.go` — xử lý xóa, một nguồn lỗi phổ biến khác
5. `x11.go`, `x11_introspector.c` / `wl_introspector.go`, `gnome_introspector.go` — lớp phát hiện ứng dụng đang gõ, quyết định dùng chế độ nào
6. `trie.go`, `mactab.go` — cấu trúc dữ liệu cho gõ tắt

**0.5 Trace giao tiếp D-Bus thực tế**

Dùng `dbus-monitor --session` khi gõ thử, để thấy chính xác engine gửi gì cho `ibus-daemon` lúc lỗi xảy ra — đặc biệt hữu ích khi so sánh giữa app chạy tốt và app bị lỗi.

### Deliverable
File `docs/known-issues.md` trong repo của bạn: mỗi lỗi còn tồn đọng có mô tả, bước tái hiện, giả thuyết nguyên nhân (dựa vào code đã đọc ở 0.4), và mức độ ưu tiên.

---

## Giai đoạn 1 — Fork & làm quen codebase (3–4 tuần)

### Mục tiêu
Có một bản fork chạy được trên máy bạn, với bộ test case bao phủ các lỗi đã xác định ở Giai đoạn 0.

### Việc cần làm

**1.1 Fork ibus-lotus, không phải ibus-bamboo gốc**
- Fork trên GitHub, clone về máy
- Đọc kỹ `go.mod` để hiểu các dependency, đặc biệt `BambooEngine/goibus` và `BambooEngine/bamboo-core` (kiểm tra xem lotus có tự fork riêng 2 thư viện này không, hay vẫn phụ thuộc vào bản gốc)

**1.2 Kiểm tra license từng phần trước khi đi xa hơn**
Xác nhận license cụ thể của `goibus` và `bamboo-core` (không chỉ ibus-bamboo/ibus-lotus) — quan trọng nếu sau này bạn cân nhắc mô hình không hoàn toàn mã nguồn mở.

**1.3 Viết test case cho từng bug ở known-issues.md**
Tham khảo cấu trúc `engine_test.go`, `utils_test.go`, `emoji_test.go` có sẵn trong repo để viết test theo đúng convention của dự án. Với các bug liên quan hành vi UI (không phải logic thuần), test tự động khó hơn — ghi lại thành checklist test tay, chạy lại sau mỗi lần sửa.

**1.4 Build & chạy thử theo quy trình CI nếu có**
Kiểm tra thư mục `.github/` xem có GitHub Actions workflow build sẵn không — nếu có, dùng làm chuẩn tham khảo cho môi trường build đúng.

### Deliverable
Repo fork build được từ source, chạy ổn định trên máy bạn với các fix hiện có của lotus, có bộ test case (tự động + checklist tay) bao phủ các lỗi từ Giai đoạn 0.

---

## Giai đoạn 2 — Sửa lỗi tồn đọng trên IBus (3–4 tuần)

### Mục tiêu
Không còn lỗi nào trong nhóm "cơ bản" (mất chữ, nhảy con trỏ, search box) trên các app bạn dùng hàng ngày, chạy trên cả X11 và Wayland.

### Cách tiếp cận

**2.1 Ưu tiên theo nhóm nguyên nhân, không theo từng app riêng lẻ**
Nhiều lỗi bề ngoài khác nhau (mất chữ trên app A, nhảy con trỏ trên app B) có thể cùng gốc ở logic pre-edit/surrounding-text trong `engine_preedit.go`. Sửa đúng gốc sẽ tự động fix nhiều lỗi bề ngoài cùng lúc.

**2.2 Công cụ debug**
- Phần Go: dùng `delve` (`dlv debug`) để breakpoint vào `engine.go`, `engine_preedit.go`
- Phần C (x11_*.c): dùng `gdb` nếu lỗi liên quan tới bắt sự kiện chuột/clipboard qua Xlib
- Log tạm thời: thêm `log.Printf` tại các điểm quyết định chế độ gõ (pre-edit vs surrounding-text) để xem engine đang chọn nhánh nào khi lỗi xảy ra

**2.3 Checklist app cần test sau mỗi fix**
- GNOME Text Editor / gedit
- LibreOffice Writer, Calc
- Firefox: cả nội dung trang và address bar/search box
- GNOME Files (search box là nơi bamboo gốc bị lỗi theo báo cáo cộng đồng)

**2.4 Cân nhắc gửi PR ngược lại ibus-lotus**
Nếu fix của bạn mang tính tổng quát (không chỉ phục vụ nhu cầu riêng), gửi Pull Request về ibus-lotus — vừa tăng khả năng được review bởi người đã quen codebase, vừa tránh phân mảnh cộng đồng thêm một fork nữa.

### Deliverable
Video/ghi chú demo gõ ổn định trên 4 app checklist trên, cả X11 và Wayland. Known-issues.md cập nhật trạng thái "đã fix" cho từng mục.

---

## Giai đoạn 3 — Fcitx5 addon (3–5 tuần)

### Mục tiêu
Bộ gõ chạy được trên Fcitx5, ưu tiên trải nghiệm tốt trên Wayland — đây là phần hoàn toàn mới, không có sẵn để fork.

### Việc cần làm

**3.1 Nghiên cứu fcitx5-unikey làm tài liệu tham khảo cấu trúc**
Đây là addon Fcitx5 tiếng Việt đã tồn tại — dù không dùng chung thuật toán, cấu trúc project (CMakeLists.txt, cách đăng ký addon, cách implement class InputMethodEngine) là tài liệu sống hữu ích nhất để học Fcitx5 addon API.

**3.2 Cài môi trường phát triển Fcitx5**
```
sudo apt install fcitx5 fcitx5-frontend-all libfcitx5core-dev cmake extra-cmake-modules
```
(Tên gói cụ thể có thể khác theo distro/version — đối chiếu tài liệu chính thức fcitx.github.io lúc thực hiện.)

**3.3 Xuất bamboo-core ra C ABI qua Go**
```go
// package libvicore, build với: go build -buildmode=c-shared -o libvicore.so
package main

import "C"
import bamboo "your-fork/bamboo-core"

//export ProcessKey
func ProcessKey(keyCode C.int, state C.int) *C.char {
    // gọi vào bamboo-core, trả về composing string hoặc commit string
    result := bamboo.HandleKey(...)
    return C.CString(result)
}

func main() {} // bắt buộc phải có khi build c-shared
```
Đây là khung sườn minh hoạ — cấu trúc API thật cần thiết kế dựa trên các hàm xử lý phím hiện có trong `engine.go` của bamboo-core, không phải viết mới từ đầu.

**3.4 Viết C++ addon gọi vào thư viện trên**
- Implement class kế thừa `InputMethodEngineV3` (hoặc phiên bản tương đương ở thời điểm bạn code — kiểm tra API hiện hành trong tài liệu Fcitx5)
- Include header `.h` tự sinh ra từ bước `-buildmode=c-shared`, gọi các hàm `//export`
- Đăng ký addon qua file `.conf` theo chuẩn Fcitx5

**3.5 Ưu tiên `text-input-v3` cho Wayland ngay từ đầu**
Vì cả README ibus-bamboo lẫn HACKING.adoc đều xác nhận IBus trên Wayland còn yếu, đây chính là chỗ Fcitx5 addon có thể tạo giá trị khác biệt rõ nhất — đầu tư đúng mức ngay từ giai đoạn này thay vì để dành "tối ưu sau".

### Deliverable
Addon Fcitx5 build được, gõ ổn định trên KDE Plasma + Wayland, dùng chế độ surrounding-text làm chính (không phụ thuộc pre-edit gạch chân).

---

## Giai đoạn 4 — Electron apps hardening (3–4 tuần)

### Mục tiêu
5 ứng dụng dùng nhiều nhất ở dân văn phòng Việt Nam gõ mượt, không mất chữ, không lệch vị trí.

### Danh sách ứng dụng ưu tiên
1. Chrome/Chromium (browser dùng nhiều nhất)
2. VSCode
3. Zalo PC (Electron)
4. LibreOffice (không phải Electron nhưng vẫn cần re-test sau các thay đổi ở Giai đoạn 2–3)
5. Một app chat web (Slack hoặc Microsoft Teams qua trình duyệt)

### Kỹ thuật debug riêng cho Electron
- Chrome/Electron hỗ trợ cờ `--enable-logging --v=1` khi khởi chạy để log chi tiết hành vi input
- Dùng `chrome://inspect` hoặc DevTools của chính app Electron để xem console log khi gõ, phát hiện app có đang can thiệp vào composition event theo cách không chuẩn không
- So sánh hành vi giữa chế độ pre-edit và surrounding-text trên từng app — ghi lại app nào cần chế độ nào (giống cách ibus-bamboo cho phép chọn chế độ riêng theo từng ứng dụng)

### Deliverable
Bảng ghi chú per-app: chế độ gõ khuyến nghị, quirk riêng, cách xử lý (nếu cần cấu hình đặc biệt, đưa vào GUI cấu hình ở Giai đoạn 5 dưới dạng preset).

---

## Giai đoạn 5 — GUI cấu hình + đóng gói (3–4 tuần)

### Mục tiêu
Cài đặt và cấu hình không cần sửa file text hay chạy lệnh terminal.

### Việc cần làm

**5.1 GUI cấu hình**
- Fcitx5 đã có framework cấu hình riêng (fcitx5-configtool, hoặc KCM module cho KDE) — tận dụng thay vì viết GUI riêng từ đầu nếu có thể
- Với IBus, tham khảo cách ibus-bamboo/lotus hiện đang expose property menu, mở rộng thêm phần cấu hình kiểu gõ/gõ tắt qua GUI thay vì file text

**5.2 Đóng gói .deb**
```
# Cấu trúc cơ bản, chi tiết theo chuẩn Debian packaging
debian/
├── control
├── rules
├── changelog
└── install
```
Dùng `dpkg-buildpackage` hoặc `debuild` để build gói cài đặt 1 lệnh.

**5.3 Đóng gói Flatpak (tuỳ chọn, ưu tiên thấp hơn .deb ở giai đoạn này)**
Input method đóng gói Flatpak có một số giới hạn kỹ thuật riêng (sandbox, D-Bus permission) — cân nhắc kỹ trước khi đầu tư, có thể để lại cho giai đoạn sau MVP.

### Deliverable
File `.deb` cài đặt bằng `sudo dpkg -i`, tự động cấu hình IBus/Fcitx5 làm bộ gõ mặc định, có GUI để đổi kiểu gõ và gõ tắt.

---

## Giai đoạn 6 — Dogfood mở rộng (liên tục)

### Mục tiêu
Thu thập đủ dữ liệu sử dụng thật để quyết định bước tiếp theo — mở rộng thành sản phẩm, đóng góp ngược cho ibus-lotus, hay giữ như một công cụ cá nhân.

### Việc cần làm
- Tự dùng làm bộ gõ chính hàng ngày, ghi log các lần gõ lỗi (đơn giản nhất: một phím tắt để tự báo "vừa gõ lỗi" và tự động lưu lại ngữ cảnh)
- Mời 3–5 người quen (ưu tiên người đang thực sự dùng Linux vì lý do bản quyền) dùng thử, thu phản hồi qua GitHub Issues
- Liên hệ maintainer ibus-lotus để trao đổi định hướng — tránh hai bên phát triển trùng lặp, cân nhắc khả năng hợp nhất công sức
- Theo dõi diễn biến chính sách bản quyền phần mềm — đây là yếu tố bên ngoài ảnh hưởng trực tiếp đến quy mô nhu cầu thực tế cho công cụ này

### Điểm quyết định
Sau khoảng 4–6 tuần dogfood, tự đánh giá lại: mức độ ổn định đã đủ để giới thiệu rộng hơn chưa, có đáng đầu tư thêm thời gian/công sức để làm sản phẩm nghiêm túc hơn không, hay giữ nguyên như một dự án cá nhân hữu ích là đủ.
