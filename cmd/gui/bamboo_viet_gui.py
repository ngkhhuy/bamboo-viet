#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Bamboo Viet - UniKey Style Control Panel for Linux
Modern Vietnamese IME Control Panel built with GTK3
"""

import os
import sys
import json
import subprocess
import gi

gi.require_version('Gtk', '3.0')
from gi.repository import Gtk, Gdk, GLib

CONFIG_DIR = os.path.expanduser("~/.config/ibus-bamboo")
CONFIG_FILE = os.path.join(CONFIG_DIR, "ibus-bamboo.config.json")
MACRO_FILE = os.path.join(CONFIG_DIR, "ibus-bamboo.macro.text")

# Input method definitions
INPUT_METHODS = ["Telex", "VNI", "Simple Telex"]
CHARSETS = ["Unicode", "TCVN3 (ABC)", "VNI Windows", "Unicode Tổ Hợp", "VIQR"]
INPUT_MODES = [
    (1, "1. Pre-edit (Có gạch chân)"),
    (2, "2. Surrounding Text (Xóa lùi thông minh - Không gạch chân)"),
    (3, "3. ForwardKeyEvent I"),
    (4, "4. ForwardKeyEvent II"),
    (5, "5. Forward as Commit (Tối ưu Terminal)"),
    (6, "6. XTestFakeKeyEvent"),
]

class DonateDialog(Gtk.Dialog):
    def __init__(self, parent):
        super().__init__(title="☕ Ủng Hộ Tác Giả - Bamboo Viet", transient_for=parent, flags=0)
        self.set_default_size(480, 320)
        self.set_modal(True)
        self.set_resizable(False)

        box = self.get_content_area()
        box.set_spacing(12)
        box.set_margin_start(20)
        box.set_margin_end(20)
        box.set_margin_top(20)
        box.set_margin_bottom(15)

        lbl_title = Gtk.Label(label="<span size='large' weight='bold'>🎋 Cảm Ơn Bạn Đã Sử Dụng Bamboo Viet!</span>")
        lbl_title.set_use_markup(True)
        lbl_title.set_xalign(0)
        box.pack_start(lbl_title, False, False, 0)

        lbl_desc = Gtk.Label(
            label="Bamboo Viet là dự án phần mềm mã nguồn mở hoàn toàn miễn phí.\n"
                  "Sự ủng hộ tự nguyện từ bạn là nguồn động viên quý giá giúp tác giả duy trì máy chủ, nghiên cứu cải tiến và phát triển bộ gõ ngày càng hoàn thiện hơn trên Linux!"
        )
        lbl_desc.set_line_wrap(True)
        lbl_desc.set_xalign(0)
        box.pack_start(lbl_desc, False, False, 0)

        # Donation Info Frame
        frame = Gtk.Frame(label=" Các Kênh Ủng Hộ Tự Nguyện ")
        frame.set_shadow_type(Gtk.ShadowType.ETCHED_IN)
        fgrid = Gtk.Grid()
        fgrid.set_row_spacing(10)
        fgrid.set_column_spacing(15)
        fgrid.set_margin_start(15)
        fgrid.set_margin_end(15)
        fgrid.set_margin_top(15)
        fgrid.set_margin_bottom(15)
        frame.add(fgrid)
        box.pack_start(frame, True, True, 0)

        # Row 1: Ko-fi
        lbl_kofi = Gtk.Label(label="☕ <b>Ủng Hộ Qua Ko-fi:</b>")
        lbl_kofi.set_use_markup(True)
        lbl_kofi.set_xalign(0)
        fgrid.attach(lbl_kofi, 0, 0, 1, 1)

        btn_kofi = Gtk.LinkButton(uri="https://ko-fi.com/ngkhhuy", label="ko-fi.com/ngkhhuy")
        fgrid.attach(btn_kofi, 1, 0, 1, 1)

        # Row 2: GitHub Repository
        lbl_gh = Gtk.Label(label="⭐ <b>GitHub Project:</b>")
        lbl_gh.set_use_markup(True)
        lbl_gh.set_xalign(0)
        fgrid.attach(lbl_gh, 0, 1, 1, 1)

        btn_gh = Gtk.LinkButton(uri="https://github.com/ngkhhuy/bamboo-viet", label="github.com/ngkhhuy/bamboo-viet")
        fgrid.attach(btn_gh, 1, 1, 1, 1)

        self.add_button("Đóng", Gtk.ResponseType.CLOSE)
        self.show_all()


class MacroEditorDialog(Gtk.Dialog):
    def __init__(self, parent):
        super().__init__(title="Bamboo Viet - Bảng Gõ Tắt", transient_for=parent, flags=0)
        self.set_default_size(500, 400)
        self.set_modal(True)

        box = self.get_content_area()
        box.set_spacing(10)
        box.set_margin_start(15)
        box.set_margin_end(15)
        box.set_margin_top(15)
        box.set_margin_bottom(15)

        lbl = Gtk.Label(label="<b>Định dạng gõ tắt:</b> <i>từ_viết_tắt : cụm_từ_thay_thế</i> (mỗi dòng một từ)")
        lbl.set_use_markup(True)
        lbl.set_xalign(0)
        box.pack_start(lbl, False, False, 0)

        # Scrolled Text View
        scrolled = Gtk.ScrolledWindow()
        scrolled.set_hexpand(True)
        scrolled.set_vexpand(True)
        self.text_view = Gtk.TextView()
        self.text_view.set_monospace(True)
        scrolled.add(self.text_view)
        box.pack_start(scrolled, True, True, 0)

        # Load existing macro text
        self.load_macros()

        # Action Buttons
        self.add_button("Đóng", Gtk.ResponseType.CANCEL)
        btn_save = self.add_button("Lưu Danh Sách", Gtk.ResponseType.OK)
        btn_save.get_style_context().add_class("suggested-action")

        self.show_all()

    def load_macros(self):
        buf = self.text_view.get_buffer()
        if os.path.exists(MACRO_FILE):
            try:
                with open(MACRO_FILE, "r", encoding="utf-8") as f:
                    buf.set_text(f.read())
            except Exception as e:
                buf.set_text(f"# Lỗi đọc file gõ tắt: {e}")
        else:
            default_macro = "# Bảng gõ tắt mẫu Bamboo Viet\nvn:Việt Nam\nhcm:Hồ Chí Minh\nhn:Hà Nội\n"
            buf.set_text(default_macro)

    def save_macros(self):
        buf = self.text_view.get_buffer()
        start, end = buf.get_bounds()
        content = buf.get_text(start, end, True)
        os.makedirs(CONFIG_DIR, exist_ok=True)
        with open(MACRO_FILE, "w", encoding="utf-8") as f:
            f.write(content)


class UniKeyControlPanel(Gtk.Window):
    def __init__(self):
        super().__init__(title="Bamboo Viet - Bảng Điều Khiển")
        self.set_default_size(490, 250)
        self.set_resizable(False)
        self.set_position(Gtk.WindowPosition.CENTER)

        # Set window icon if exists
        icon_path = "/usr/share/ibus-bamboo/icons/vi.svg"
        if not os.path.exists(icon_path):
            icon_path = os.path.join(os.path.dirname(__file__), "../../icons/vi.svg")
        if os.path.exists(icon_path):
            self.set_icon_from_file(icon_path)

        self.config = self.load_config()

        # Main Layout Box
        main_vbox = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=10)
        main_vbox.set_margin_start(15)
        main_vbox.set_margin_end(15)
        main_vbox.set_margin_top(15)
        main_vbox.set_margin_bottom(15)
        self.add(main_vbox)

        # Top section: Main Frame (Left) and Buttons (Right)
        top_hbox = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=15)
        main_vbox.pack_start(top_hbox, False, False, 0)

        # --- LEFT FRAME: Main Controls (UniKey style) ---
        main_frame = Gtk.Frame(label=" Điều khiển chính ")
        main_frame.set_shadow_type(Gtk.ShadowType.ETCHED_IN)
        frame_grid = Gtk.Grid()
        frame_grid.set_row_spacing(10)
        frame_grid.set_column_spacing(15)
        frame_grid.set_margin_start(15)
        frame_grid.set_margin_end(15)
        frame_grid.set_margin_top(15)
        frame_grid.set_margin_bottom(15)
        main_frame.add(frame_grid)
        top_hbox.pack_start(main_frame, True, True, 0)

        # 1. Bảng mã
        lbl_charset = Gtk.Label(label="Bảng mã:")
        lbl_charset.set_xalign(0)
        frame_grid.attach(lbl_charset, 0, 0, 1, 1)

        self.combo_charset = Gtk.ComboBoxText()
        for cs in CHARSETS:
            self.combo_charset.append_text(cs)
        cur_cs = self.config.get("OutputCharset", "Unicode")
        self.combo_charset.set_active(CHARSETS.index(cur_cs) if cur_cs in CHARSETS else 0)
        frame_grid.attach(self.combo_charset, 1, 0, 1, 1)

        # 2. Kiểu gõ
        lbl_method = Gtk.Label(label="Kiểu gõ:")
        lbl_method.set_xalign(0)
        frame_grid.attach(lbl_method, 0, 1, 1, 1)

        self.combo_method = Gtk.ComboBoxText()
        for im in INPUT_METHODS:
            self.combo_method.append_text(im)
        cur_im = self.config.get("InputMethod", "Telex")
        self.combo_method.set_active(INPUT_METHODS.index(cur_im) if cur_im in INPUT_METHODS else 0)
        frame_grid.attach(self.combo_method, 1, 1, 1, 1)

        # 3. Chế độ gõ mặc định
        lbl_mode = Gtk.Label(label="Chế độ gõ:")
        lbl_mode.set_xalign(0)
        frame_grid.attach(lbl_mode, 0, 2, 1, 1)

        self.combo_mode = Gtk.ComboBoxText()
        for mid, mname in INPUT_MODES:
            self.combo_mode.append_text(mname)
        cur_mode = self.config.get("DefaultInputMode", 1)
        mode_idx = 0
        for i, (mid, _) in enumerate(INPUT_MODES):
            if mid == cur_mode:
                mode_idx = i
                break
        self.combo_mode.set_active(mode_idx)
        frame_grid.attach(self.combo_mode, 1, 2, 1, 1)

        # --- RIGHT: Action Buttons (UniKey Style) ---
        btn_vbox = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=8)
        top_hbox.pack_start(btn_vbox, False, False, 0)

        # Button Đóng
        btn_close = Gtk.Button(label="Đóng")
        btn_close.set_size_request(115, 30)
        btn_close.get_style_context().add_class("suggested-action")
        btn_close.connect("clicked", self.on_close_clicked)
        btn_vbox.pack_start(btn_close, False, False, 0)

        # Button Mở rộng / Thu gọn
        self.btn_expand = Gtk.Button(label="Mở rộng >>")
        self.btn_expand.set_size_request(115, 30)
        self.btn_expand.connect("clicked", self.on_toggle_expand)
        btn_vbox.pack_start(self.btn_expand, False, False, 0)

        # Button Bảng gõ tắt
        btn_macro = Gtk.Button(label="Bảng gõ tắt...")
        btn_macro.set_size_request(115, 30)
        btn_macro.connect("clicked", self.on_open_macro_editor)
        btn_vbox.pack_start(btn_macro, False, False, 0)

        # Button Mặc định
        btn_default = Gtk.Button(label="Mặc định")
        btn_default.set_size_request(115, 30)
        btn_default.connect("clicked", self.on_reset_default)
        btn_vbox.pack_start(btn_default, False, False, 0)

        # Button Ủng hộ tác giả (Donate)
        btn_donate = Gtk.Button(label="☕ Ủng hộ...")
        btn_donate.set_size_request(115, 30)
        btn_donate.connect("clicked", self.on_donate_clicked)
        btn_vbox.pack_start(btn_donate, False, False, 0)

        # Button Thông tin
        btn_about = Gtk.Button(label="Thông tin...")
        btn_about.set_size_request(115, 30)
        btn_about.connect("clicked", self.on_about_clicked)
        btn_vbox.pack_start(btn_about, False, False, 0)

        # --- EXPANDABLE ADVANCED OPTIONS FRAME ---
        self.adv_frame = Gtk.Frame(label=" Tùy chọn nâng cao ")
        self.adv_frame.set_shadow_type(Gtk.ShadowType.ETCHED_IN)
        self.adv_frame.set_no_show_all(True)

        adv_grid = Gtk.Grid()
        adv_grid.set_row_spacing(8)
        adv_grid.set_column_spacing(20)
        adv_grid.set_margin_start(15)
        adv_grid.set_margin_end(15)
        adv_grid.set_margin_top(10)
        adv_grid.set_margin_bottom(10)
        self.adv_frame.add(adv_grid)

        # Checkboxes
        self.chk_spell = Gtk.CheckButton(label="Bật kiểm tra chính tả")
        self.chk_spell.set_active(True)
        adv_grid.attach(self.chk_spell, 0, 0, 1, 1)

        self.chk_restore = Gtk.CheckButton(label="Tự động khôi phục phím với từ sai")
        self.chk_restore.set_active(True)
        adv_grid.attach(self.chk_restore, 1, 0, 1, 1)

        self.chk_new_tone = Gtk.CheckButton(label="Đặt dấu kiểu mới (hòa, thủy)")
        self.chk_new_tone.set_active(True)
        adv_grid.attach(self.chk_new_tone, 0, 1, 1, 1)

        self.chk_free_typing = Gtk.CheckButton(label="Cho phép gõ tự do (Free typing)")
        self.chk_free_typing.set_active(True)
        adv_grid.attach(self.chk_free_typing, 1, 1, 1, 1)

        self.chk_chat_fix = Gtk.CheckButton(label="Sửa lỗi lặp từ Enter trong Chat (Zalo/Slack)")
        self.chk_chat_fix.set_active(True)
        adv_grid.attach(self.chk_chat_fix, 0, 2, 1, 1)

        self.chk_enable_macro = Gtk.CheckButton(label="Bật tính năng gõ tắt")
        self.chk_enable_macro.set_active(True)
        adv_grid.attach(self.chk_enable_macro, 1, 2, 1, 1)

        main_vbox.pack_start(self.adv_frame, False, False, 0)

        # Connect signals
        self.connect("destroy", Gtk.main_quit)
        self.is_expanded = False

    def load_config(self):
        if os.path.exists(CONFIG_FILE):
            try:
                with open(CONFIG_FILE, "r", encoding="utf-8") as f:
                    return json.load(f)
            except Exception:
                pass
        return {
            "InputMethod": "Telex",
            "OutputCharset": "Unicode",
            "DefaultInputMode": 1,
            "Flags": 7,
            "InputModeMapping": {}
        }

    def save_config(self):
        cs_idx = self.combo_charset.get_active()
        if cs_idx >= 0 and cs_idx < len(CHARSETS):
            self.config["OutputCharset"] = CHARSETS[cs_idx]

        im_idx = self.combo_method.get_active()
        if im_idx >= 0 and im_idx < len(INPUT_METHODS):
            self.config["InputMethod"] = INPUT_METHODS[im_idx]

        mode_idx = self.combo_mode.get_active()
        if mode_idx >= 0 and mode_idx < len(INPUT_MODES):
            self.config["DefaultInputMode"] = INPUT_MODES[mode_idx][0]

        os.makedirs(CONFIG_DIR, exist_ok=True)
        try:
            with open(CONFIG_FILE, "w", encoding="utf-8") as f:
                json.dump(self.config, f, indent=2, ensure_ascii=False)
            # Notify ibus daemon if running
            subprocess.run(["ibus", "restart"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        except Exception as e:
            print(f"Error saving config: {e}")

    def on_toggle_expand(self, widget):
        self.is_expanded = not self.is_expanded
        if self.is_expanded:
            self.adv_frame.show_all()
            self.btn_expand.set_label("<< Thu gọn")
        else:
            self.adv_frame.hide()
            self.btn_expand.set_label("Mở rộng >>")
        self.resize(1, 1)

    def on_open_macro_editor(self, widget):
        dialog = MacroEditorDialog(self)
        response = dialog.run()
        if response == Gtk.ResponseType.OK:
            dialog.save_macros()
        dialog.destroy()

    def on_donate_clicked(self, widget):
        dialog = DonateDialog(self)
        dialog.run()
        dialog.destroy()

    def on_reset_default(self, widget):
        self.combo_charset.set_active(0)
        self.combo_method.set_active(0)
        self.combo_mode.set_active(0)
        self.chk_spell.set_active(True)
        self.chk_restore.set_active(True)
        self.chk_new_tone.set_active(True)
        self.chk_free_typing.set_active(True)
        self.chk_chat_fix.set_active(True)
        self.chk_enable_macro.set_active(True)

    def on_about_clicked(self, widget):
        about = Gtk.AboutDialog(transient_for=self, modal=True)
        about.set_program_name("Bamboo Viet")
        about.set_version("1.0.0")
        about.set_comments("Bộ gõ tiếng Việt hiện đại cho Linux (hỗ trợ Wayland native, Fcitx5 và IBus).\nĐã khắc phục triệt để lỗi Enter trong Chat, LibreOffice, Search Box và Electron apps.")
        about.set_website("https://github.com/ngkhhuy/bamboo-viet")
        about.set_website_label("GitHub Repository")
        about.set_authors(["Bamboo Viet Project Contributors", "Luong Thanh Lam (Original author)"])
        about.set_license_type(Gtk.License.GPL_3_0)
        about.run()
        about.destroy()

    def on_close_clicked(self, widget):
        self.save_config()
        self.destroy()
        Gtk.main_quit()


def main():
    app = UniKeyControlPanel()
    app.show_all()
    Gtk.main()

if __name__ == "__main__":
    main()
