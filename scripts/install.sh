#!/usr/bin/env bash
set -e

# Bamboo Viet One-Line Installer
# Repository: https://github.com/ngkhhuy/bamboo-viet

BOLD='\033[1m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BOLD}${BLUE}🎋 Đang cài đặt Bamboo Viet — Bộ gõ tiếng Việt cho Linux...${NC}"

# Check sudo access
if [ "$EUID" -ne 0 ]; then
    SUDO="sudo"
else
    SUDO=""
fi

PKG_NAME="ibus-bamboo-viet_1.0.1_amd64.deb"
TEMP_DIR=$(mktemp -d)
CLEANUP=true

# Cleanup on exit
trap 'if [ "$CLEANUP" = true ]; then rm -rf "$TEMP_DIR"; fi' EXIT

DEB_PATH=""

# 1. Check if running inside bamboo-viet source tree with pre-built deb
if [ -f "bin/$PKG_NAME" ]; then
    echo -e "${GREEN}✓ Tìm thấy file .deb có sẵn trong thư mục bin/...${NC}"
    DEB_PATH="bin/$PKG_NAME"
    CLEANUP=false
elif [ -f "$PKG_NAME" ]; then
    DEB_PATH="$PKG_NAME"
    CLEANUP=false
else
    echo -e "${BLUE}⬇ Đang tải gói cài đặt từ GitHub Releases...${NC}"
    RELEASE_URL="https://github.com/ngkhhuy/bamboo-viet/releases/latest/download/$PKG_NAME"
    
    if wget -q --spider "$RELEASE_URL" 2>/dev/null; then
        wget -q --show-progress -O "$TEMP_DIR/$PKG_NAME" "$RELEASE_URL"
        DEB_PATH="$TEMP_DIR/$PKG_NAME"
    else
        echo -e "${YELLOW}ℹ Chưa tìm thấy bản Release online. Đang tự động biên dịch từ mã nguồn...${NC}"
        # Install build dependencies if needed
        $SUDO apt-get update -qq
        $SUDO apt-get install -y -qq gcc g++ make dpkg-dev libx11-dev libxtst-dev golang git >/dev/null 2>&1 || true
        
        SRC_DIR="$TEMP_DIR/bamboo-viet"
        git clone --depth 1 https://github.com/ngkhhuy/bamboo-viet.git "$SRC_DIR"
        (cd "$SRC_DIR" && make deb)
        DEB_PATH="$SRC_DIR/bin/$PKG_NAME"
    fi
fi

# 2. Install package
echo -e "${BLUE}📦 Đang cài đặt vào hệ thống...${NC}"
$SUDO dpkg -i "$DEB_PATH" || $SUDO apt-get install -f -y

# 3. Restart IBus
echo -e "${BLUE}🔄 Đang khởi động lại IBus...${NC}"
ibus restart || echo "Lưu ý: Có thể chạy 'ibus restart' bằng tay nếu cần."
sleep 1

# 4. Auto-register in GNOME Input Sources if applicable
if command -v gsettings >/dev/null 2>&1; then
    CURRENT_SOURCES=$(gsettings get org.gnome.desktop.input-sources sources 2>/dev/null || echo "")
    if [ -n "$CURRENT_SOURCES" ] && [[ "$CURRENT_SOURCES" != *"('ibus', 'Bamboo')"* ]]; then
        # Add ('ibus', 'Bamboo') to existing sources list
        if [ "$CURRENT_SOURCES" = "@a(ss) []" ] || [ "$CURRENT_SOURCES" = "[]" ]; then
            NEW_SOURCES="[('ibus', 'Bamboo')]"
        else
            NEW_SOURCES=$(echo "$CURRENT_SOURCES" | sed "s/]$/, ('ibus', 'Bamboo')]/")
        fi
        gsettings set org.gnome.desktop.input-sources sources "$NEW_SOURCES" 2>/dev/null || true
        echo -e "${GREEN}✓ Đã tự động kích hoạt bộ gõ 'Bamboo' vào thanh ngôn ngữ GNOME.${NC}"
    fi
fi

echo -e "\n${BOLD}${GREEN}✅ Cài đặt Bamboo Viet thành công!${NC}\n"
echo -e "${BOLD}Hướng dẫn sử dụng:${NC}"
echo -e "  - ${BOLD}Chuyển đổi bộ gõ:${NC} Nhấn phím tắt ${BOLD}Super + Space${NC} (hoặc Ctrl + Space)"
echo -e "  - ${BOLD}Mở Bảng điều khiển:${NC} Chạy lệnh ${BOLD}bamboo-viet-gui${NC} hoặc tìm ${BOLD}Bamboo Viet Control Panel${NC} trong Menu ứng dụng"
echo -e "  - ${BOLD}Nếu cần thêm thủ công:${NC} Vào ${BOLD}Settings -> Keyboard -> Input Sources${NC}, nhấn ${BOLD}+${NC} -> chọn ${BOLD}Vietnamese${NC} -> chọn ${BOLD}Bamboo${NC}\n"

