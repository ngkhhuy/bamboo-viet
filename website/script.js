/**
 * Bamboo Viet — Modern Landing Page Interactive Logic
 * Features:
 *  1. Client-Side Vietnamese IME Engine (Telex & VNI) for Interactive Playground
 *  2. Distro Tab Switcher with 1-Click Terminal Command Copy
 *  3. Dark / Light Theme Toggle with LocalStorage persistence
 *  4. FAQ Accordion & Mobile Navigation Drawer
 *  5. Interactive UniKey GUI Mockup Controls
 */

document.addEventListener('DOMContentLoaded', () => {
  // =========================================================================
  // 1. Theme Management (Dark / Light)
  // =========================================================================
  const themeToggleBtn = document.getElementById('theme-toggle-btn');
  const htmlRoot = document.documentElement;

  // Read saved theme or default to dark
  const savedTheme = localStorage.getItem('bv_theme') || 'dark';
  htmlRoot.setAttribute('data-theme', savedTheme);

  if (themeToggleBtn) {
    themeToggleBtn.addEventListener('click', () => {
      const currentTheme = htmlRoot.getAttribute('data-theme');
      const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
      htmlRoot.setAttribute('data-theme', newTheme);
      localStorage.setItem('bv_theme', newTheme);
    });
  }

  // =========================================================================
  // 2. Mobile Menu Drawer
  // =========================================================================
  const mobileMenuBtn = document.getElementById('mobile-menu-btn');
  const mobileDrawer = document.getElementById('mobile-drawer');

  if (mobileMenuBtn && mobileDrawer) {
    mobileMenuBtn.addEventListener('click', () => {
      mobileDrawer.classList.toggle('open');
    });

    // Close mobile drawer when clicking a link
    mobileDrawer.querySelectorAll('a').forEach(link => {
      link.addEventListener('click', () => {
        mobileDrawer.classList.remove('open');
      });
    });
  }

  // =========================================================================
  // 3. Interactive Vietnamese IME Engine (Telex & VNI)
  // =========================================================================
  const playgroundInput = document.getElementById('playground-input');
  const charCounter = document.getElementById('char-counter');
  const playgroundClearBtn = document.getElementById('playground-clear-btn');
  const btnMethodTelex = document.getElementById('btn-method-telex');
  const btnMethodVni = document.getElementById('btn-method-vni');
  const statusText = document.getElementById('playground-status-text');

  let currentInputMethod = 'telex'; // 'telex' or 'vni'

  // Vietnamese Transformation Rules
  const vowels = {
    'a': { s: 'á', f: 'à', r: 'ả', x: 'ã', j: 'ạ', raw: 'a' },
    'ă': { s: 'ắ', f: 'ằ', r: 'ẳ', x: 'ẵ', j: 'ặ', raw: 'ă' },
    'â': { s: 'ấ', f: 'ầ', r: 'ẩ', x: 'ẫ', j: 'ậ', raw: 'â' },
    'e': { s: 'é', f: 'è', r: 'ẻ', x: 'ẽ', j: 'ẹ', raw: 'e' },
    'ê': { s: 'ế', f: 'ề', r: 'ể', x: 'ễ', j: 'ệ', raw: 'ê' },
    'i': { s: 'í', f: 'ì', r: 'ỉ', x: 'ĩ', j: 'ị', raw: 'i' },
    'o': { s: 'ó', f: 'ò', r: 'ỏ', x: 'õ', j: 'ọ', raw: 'o' },
    'ô': { s: 'ố', f: 'ồ', r: 'ổ', x: 'ỗ', j: 'ộ', raw: 'ô' },
    'ơ': { s: 'ớ', f: 'ờ', r: 'ở', x: 'ỡ', j: 'ợ', raw: 'ơ' },
    'u': { s: 'ú', f: 'ù', r: 'ủ', x: 'ũ', j: 'ụ', raw: 'u' },
    'ư': { s: 'ứ', f: 'ừ', r: 'ử', x: 'ữ', j: 'ự', raw: 'ư' },
    'y': { s: 'ý', f: 'ỳ', r: 'ỷ', x: 'ỹ', j: 'ỵ', raw: 'y' },
  };

  // VNI Tone Key Map: 1: sắc, 2: huyền, 3: hỏi, 4: ngã, 5: nặng, 0: xóa
  const vniToneMap = { '1': 's', '2': 'f', '3': 'r', '4': 'x', '5': 'j', '0': 'z' };

  function removeTone(char) {
    for (const base in vowels) {
      const obj = vowels[base];
      for (const t of ['s', 'f', 'r', 'x', 'j']) {
        if (obj[t] === char.toLowerCase()) {
          return char === char.toUpperCase() ? base.toUpperCase() : base;
        }
      }
    }
    return char;
  }

  function getToneOfChar(char) {
    for (const base in vowels) {
      const obj = vowels[base];
      for (const t of ['s', 'f', 'r', 'x', 'j']) {
        if (obj[t] === char.toLowerCase()) return t;
      }
    }
    return null;
  }

  function applyToneToVowel(vowelChar, toneKey) {
    const isUpper = vowelChar === vowelChar.toUpperCase();
    const lower = vowelChar.toLowerCase();
    const rawVowel = removeTone(lower);
    if (vowels[rawVowel]) {
      if (toneKey === 'z') return isUpper ? rawVowel.toUpperCase() : rawVowel;
      const toned = vowels[rawVowel][toneKey] || rawVowel;
      return isUpper ? toned.toUpperCase() : toned;
    }
    return vowelChar;
  }

  // Transform word using Telex
  function processWordTelex(word) {
    if (!word) return word;
    let res = word;

    // 1. Double letter conversions (dd, aa, ee, oo, aw, ow, uw)
    res = res.replace(/dd/gi, (m) => (m[0] === 'D' || m[1] === 'D' ? (m === 'DD' ? 'Đ' : 'Đ') : 'đ'));
    res = res.replace(/aa/gi, (m) => (m[0] === 'A' ? 'Â' : 'â'));
    res = res.replace(/aw/gi, (m) => (m[0] === 'A' ? 'Ă' : 'ă'));
    res = res.replace(/ee/gi, (m) => (m[0] === 'E' ? 'Ê' : 'ê'));
    res = res.replace(/oo/gi, (m) => (m[0] === 'O' ? 'Ô' : 'ô'));
    res = res.replace(/ow/gi, (m) => (m[0] === 'O' ? 'Ơ' : 'ơ'));
    res = res.replace(/uw/gi, (m) => (m[0] === 'U' ? 'Ư' : 'ư'));
    res = res.replace(/w/gi, (m) => (m === 'W' ? 'Ư' : 'ư')); // standalone w

    // 2. Tones: s (sắc), f (huyền), r (hỏi), x (ngã), j (nặng), z (xóa)
    const toneKeys = ['s', 'f', 'r', 'x', 'j', 'z'];
    let lastChar = res.slice(-1).toLowerCase();

    if (toneKeys.includes(lastChar) && res.length > 1) {
      const tone = lastChar;
      const body = res.slice(0, -1);
      
      // Find suitable vowel position in reverse
      let vowelIdx = -1;
      const chars = body.split('');
      for (let i = chars.length - 1; i >= 0; i--) {
        const raw = removeTone(chars[i].toLowerCase());
        if (vowels[raw]) {
          vowelIdx = i;
          break;
        }
      }

      if (vowelIdx !== -1) {
        chars[vowelIdx] = applyToneToVowel(chars[vowelIdx], tone);
        res = chars.join('');
      }
    }

    return res;
  }

  // Transform word using VNI
  function processWordVni(word) {
    if (!word) return word;
    let res = word;

    // VNI Vowel rules: a6 -> â, a8 -> ă, e6 -> ê, o6 -> ô, o7 -> ơ, u7 -> ư, d9 -> đ
    res = res.replace(/d9/gi, (m) => (m[0] === 'D' ? 'Đ' : 'đ'));
    res = res.replace(/a6/gi, (m) => (m[0] === 'A' ? 'Â' : 'â'));
    res = res.replace(/a8/gi, (m) => (m[0] === 'A' ? 'Ă' : 'ă'));
    res = res.replace(/e6/gi, (m) => (m[0] === 'E' ? 'Ê' : 'ê'));
    res = res.replace(/o6/gi, (m) => (m[0] === 'O' ? 'Ô' : 'ô'));
    res = res.replace(/o7/gi, (m) => (m[0] === 'O' ? 'Ơ' : 'ơ'));
    res = res.replace(/u7/gi, (m) => (m[0] === 'U' ? 'Ư' : 'ư'));

    // VNI tones: 1 (sắc), 2 (huyền), 3 (hỏi), 4 (ngã), 5 (nặng), 0 (xóa)
    const lastChar = res.slice(-1);
    if (vniToneMap[lastChar] && res.length > 1) {
      const tone = vniToneMap[lastChar];
      const body = res.slice(0, -1);
      
      let vowelIdx = -1;
      const chars = body.split('');
      for (let i = chars.length - 1; i >= 0; i--) {
        const raw = removeTone(chars[i].toLowerCase());
        if (vowels[raw]) {
          vowelIdx = i;
          break;
        }
      }

      if (vowelIdx !== -1) {
        chars[vowelIdx] = applyToneToVowel(chars[vowelIdx], tone);
        res = chars.join('');
      }
    }

    return res;
  }

  // Engine input listener
  function handlePlaygroundInput() {
    if (!playgroundInput) return;
    const text = playgroundInput.value;
    const cursor = playgroundInput.selectionStart;

    // Split words by space / newline
    const words = text.split(/(\s+)/);
    const converted = words.map(w => {
      if (/\s+/.test(w)) return w;
      return currentInputMethod === 'telex' ? processWordTelex(w) : processWordVni(w);
    }).join('');

    if (converted !== text) {
      const diff = converted.length - text.length;
      playgroundInput.value = converted;
      playgroundInput.setSelectionRange(cursor + diff, cursor + diff);
    }

    if (charCounter) {
      charCounter.textContent = playgroundInput.value.length;
    }
  }

  if (playgroundInput) {
    playgroundInput.addEventListener('input', handlePlaygroundInput);
  }

  // Method Switcher Buttons
  if (btnMethodTelex && btnMethodVni) {
    btnMethodTelex.addEventListener('click', () => {
      currentInputMethod = 'telex';
      btnMethodTelex.classList.add('active');
      btnMethodVni.classList.remove('active');
      statusText.textContent = 'Đang dùng chế độ gõ Telex (aa -> â, s -> sắc, f -> huyền...)';
      playgroundInput.focus();
    });

    btnMethodVni.addEventListener('click', () => {
      currentInputMethod = 'vni';
      btnMethodVni.classList.add('active');
      btnMethodTelex.classList.remove('active');
      statusText.textContent = 'Đang dùng chế độ gõ VNI (a6 -> â, 1 -> sắc, 2 -> huyền...)';
      playgroundInput.focus();
    });
  }

  // Quick Sample Buttons
  document.querySelectorAll('.sample-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const textToSimulate = btn.getAttribute('data-text');
      if (playgroundInput && textToSimulate) {
        playgroundInput.value = '';
        let i = 0;
        const timer = setInterval(() => {
          if (i < textToSimulate.length) {
            playgroundInput.value += textToSimulate[i];
            handlePlaygroundInput();
            i++;
          } else {
            clearInterval(timer);
          }
        }, 40);
      }
    });
  });

  // Clear Playground Button
  if (playgroundClearBtn && playgroundInput) {
    playgroundClearBtn.addEventListener('click', () => {
      playgroundInput.value = '';
      if (charCounter) charCounter.textContent = '0';
      playgroundInput.focus();
    });
  }

  // =========================================================================
  // 4. Installation Distro Snippet Switcher & Copy Command
  // =========================================================================
  const distroSnippets = {
    ubuntu: {
      title: 'Ubuntu / Debian / Linux Mint / Pop!_OS (.deb)',
      code: `# Cài đặt tự động nhanh nhất với 1 dòng lệnh:
curl -fsSL https://raw.githubusercontent.com/ngkhhuy/bamboo-viet/main/scripts/install.sh | bash

# Hoặc tải file .deb thủ công:
# wget https://github.com/ngkhhuy/bamboo-viet/releases/latest/download/ibus-bamboo-viet_1.0.0_amd64.deb
# sudo dpkg -i ibus-bamboo-viet_1.0.0_amd64.deb || sudo apt-get install -f -y
# ibus restart`
    },
    arch: {
      title: 'Arch Linux / Manjaro / EndeavourOS (AUR / PKGBUILD)',
      code: `# Cách 1: Cài đặt trực tiếp qua AUR Helper (yay hoặc paru)
yay -S bamboo-viet

# Cách 2: Hoặc tự build từ PKGBUILD
git clone https://aur.archlinux.org/bamboo-viet.git
cd bamboo-viet
makepkg -si

# Khởi động lại IBus hoặc Fcitx5
ibus restart # hoặc fcitx5 -r -d`
    },
    fedora: {
      title: 'Fedora / RHEL / openSUSE',
      code: `# Cài đặt các gói phụ thuộc cần thiết
sudo dnf install -y golang ibus-devel libX11-devel libXtst-devel

# Clone repository và cài đặt
git clone https://github.com/ngkhhuy/bamboo-viet.git
cd bamboo-viet
make && sudo make install

# Khởi động lại IBus
ibus restart`
    },
    source: {
      title: 'Biên dịch từ mã nguồn (Toàn bộ Distro Linux)',
      code: `# 1. Clone mã nguồn
git clone https://github.com/ngkhhuy/bamboo-viet.git
cd bamboo-viet

# 2. Kiểm tra và biên dịch
make test
make

# 3. Cài đặt vào hệ thống
sudo make install
ibus restart`
    }
  };

  const distroTabBtns = document.querySelectorAll('.distro-tab-btn');
  const terminalDistroTitle = document.getElementById('terminal-distro-title');
  const terminalCodeBlock = document.getElementById('terminal-code-block');
  const copyCmdBtn = document.getElementById('copy-cmd-btn');
  const copyBtnText = document.getElementById('copy-btn-text');
  const toast = document.getElementById('toast');
  const toastMessage = document.getElementById('toast-message');

  function showToast(msg) {
    if (!toast) return;
    if (toastMessage) toastMessage.textContent = msg;
    toast.classList.add('show');
    setTimeout(() => {
      toast.classList.remove('show');
    }, 3000);
  }

  distroTabBtns.forEach(btn => {
    btn.addEventListener('click', () => {
      distroTabBtns.forEach(b => b.classList.remove('active'));
      btn.classList.add('active');

      const distro = btn.getAttribute('data-distro');
      const snippet = distroSnippets[distro];
      if (snippet && terminalDistroTitle && terminalCodeBlock) {
        terminalDistroTitle.textContent = snippet.title;
        terminalCodeBlock.textContent = snippet.code;
      }
    });
  });

  if (copyCmdBtn && terminalCodeBlock) {
    copyCmdBtn.addEventListener('click', () => {
      const code = terminalCodeBlock.textContent;
      navigator.clipboard.writeText(code).then(() => {
        if (copyBtnText) copyBtnText.textContent = 'Đã sao chép!';
        showToast('Đã sao chép toàn bộ lệnh vào bộ nhớ tạm!');
        setTimeout(() => {
          if (copyBtnText) copyBtnText.textContent = 'Sao chép lệnh';
        }, 2000);
      }).catch(() => {
        showToast('Không thể sao chép tự động. Hãy bôi đen để copy.');
      });
    });
  }

  // =========================================================================
  // 5. GUI Control Panel Mockup Interactive Triggers
  // =========================================================================
  const guiBtnSave = document.getElementById('gui-btn-save');
  const guiBtnExpand = document.getElementById('gui-btn-expand');
  const guiBtnCoffee = document.getElementById('gui-btn-coffee');

  if (guiBtnSave) {
    guiBtnSave.addEventListener('click', () => {
      showToast('🎋 Đã lưu cấu hình Bamboo Viet thành công!');
    });
  }

  if (guiBtnExpand) {
    guiBtnExpand.addEventListener('click', () => {
      showToast('⚙️ Bảng tùy chọn mở rộng: Phím tắt, Macro, App Presets...');
    });
  }

  if (guiBtnCoffee) {
    guiBtnCoffee.addEventListener('click', () => {
      showToast('☕ Cảm ơn bạn đã quan tâm và ủng hộ Bamboo Viet!');
    });
  }

  // =========================================================================
  // 6. FAQ Accordion Toggle
  // =========================================================================
  document.querySelectorAll('.faq-question').forEach(btn => {
    btn.addEventListener('click', () => {
      const item = btn.parentElement;
      const isActive = item.classList.contains('active');

      // Close all items
      document.querySelectorAll('.faq-item').forEach(i => i.classList.remove('active'));

      // Toggle clicked item
      if (!isActive) {
        item.classList.add('active');
      }
    });
  });
});
