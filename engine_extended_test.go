package main

import (
	"ibus-bamboo/config"
	"testing"

	"github.com/BambooEngine/bamboo-core"
	"github.com/godbus/dbus/v5"
)

// TestIsValidStateHotkeys verifies BUG-07: ensure system and app hotkeys with modifiers
// (Ctrl, Alt, Super, Meta, Hyper) are not consumed by the Vietnamese engine.
func TestIsValidStateHotkeys(t *testing.T) {
	testCases := []struct {
		name     string
		state    uint32
		expected bool
	}{
		{"NormalKey", 0, true},
		{"ShiftKey", IBusShiftMask, true},
		{"CapsLock", IBusLockMask, true},
		{"CtrlKey", IBusControlMask, false},
		{"CtrlShift", IBusControlMask | IBusShiftMask, false},
		{"AltKey_Mod1", IBusMod1Mask, false},
		{"SuperKey_Mod4", IBusMod4Mask, false},
		{"SuperKey_Super", IBusSuperMask, false},
		{"MetaKey", IBusMetaMask, false},
		{"HyperKey", IBusHyperMask, false},
		{"IgnoredKey", IBusIgnoredMask, false},
		{"CtrlAlt", IBusControlMask | IBusMod1Mask, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := isValidState(tc.state)
			if actual != tc.expected {
				t.Errorf("isValidState(%#x) = %v; want %v", tc.state, actual, tc.expected)
			}
		})
	}
}

// TestTelexTypingSequence tests complex Vietnamese syllable parsing using bamboo-core
func TestTelexTypingSequence(t *testing.T) {
	imDefs := bamboo.GetInputMethodDefinitions()
	telexMethod := bamboo.ParseInputMethod(imDefs, "Telex")

	testWords := []struct {
		input    string
		expected string
	}{
		{"tieengs", "tiếng"},
		{"vieetj", "việt"},
		{"dduwowngf", "đường"},
		{"nguyeen", "nguyên"},
		{"nguyeexn", "nguyễn"},
		{"thuyeenf", "thuyền"},
		{"hoaf", "hòa"},
		{"toanf", "toàn"},
		{"nghieengf", "nghiềng"},
		{"khuyeens", "khuyến"},
		{"quoocs", "quốc"},
		{"truwowngf", "trường"},
		{"phuwowng", "phương"},
		{"ddax", "đã"},
		{"cuwaj", "cựa"},
		{"cow", "cơ"},
		{"cowi", "cơi"},
	}

	for _, tw := range testWords {
		t.Run("Telex_"+tw.input, func(t *testing.T) {
			engine := bamboo.NewEngine(telexMethod, bamboo.EstdFlags)
			for _, r := range tw.input {
				engine.ProcessKey(r, bamboo.VietnameseMode)
			}
			result := engine.GetProcessedString(bamboo.VietnameseMode)
			if result != tw.expected {
				t.Errorf("Input '%s' produced '%s'; want '%s'", tw.input, result, tw.expected)
			}
		})
	}
}

// TestVNITypingSequence tests VNI input mode parsing
func TestVNITypingSequence(t *testing.T) {
	imDefs := bamboo.GetInputMethodDefinitions()
	vniMethod := bamboo.ParseInputMethod(imDefs, "VNI")

	testWords := []struct {
		input    string
		expected string
	}{
		{"tie6ng1", "tiếng"},
		{"vie6t5", "việt"},
		{"d9u7o7ng2", "đường"},
		{"nguye6n", "nguyên"},
		{"thuye6n2", "thuyền"},
		{"toa3n", "toản"},
		{"toa4n", "toãn"},
		{"quo61c", "quốc"},
	}

	for _, tw := range testWords {
		t.Run("VNI_"+tw.input, func(t *testing.T) {
			engine := bamboo.NewEngine(vniMethod, bamboo.EstdFlags)
			for _, r := range tw.input {
				engine.ProcessKey(r, bamboo.VietnameseMode)
			}
			result := engine.GetProcessedString(bamboo.VietnameseMode)
			if result != tw.expected {
				t.Errorf("VNI Input '%s' produced '%s'; want '%s'", tw.input, result, tw.expected)
			}
		})
	}
}

// TestWordBreakSymbols verifies bamboo word-break detection
func TestWordBreakSymbols(t *testing.T) {
	symbols := []rune{' ', '.', ',', ';', ':', '!', '?', '-', '(', ')', '[', ']', '{', '}', '/', '\\', '"', '\'', '0', '5', '9'}
	for _, sym := range symbols {
		if !bamboo.IsWordBreakSymbol(sym) {
			t.Errorf("Rune '%c' (%d) should be recognized as WordBreakSymbol", sym, sym)
		}
	}

	nonSymbols := []rune{'a', 'b', 'c', 'e', 'o', 'u', 'i', 'd', 'w', 's', 'f', 'r', 'x', 'j'}
	for _, nonSym := range nonSymbols {
		if bamboo.IsWordBreakSymbol(nonSym) {
			t.Errorf("Rune '%c' (%d) should NOT be recognized as WordBreakSymbol", nonSym, nonSym)
		}
	}
}

// TestGetLastWordFromSentence verifies helper function used in Search Box & Address Bar
func TestGetLastWordFromSentence(t *testing.T) {
	testCases := []struct {
		sentence string
		expected string
	}{
		{"", ""},
		{"   ", ""},
		{"hello", "hello"},
		{"hello world", "world"},
		{"https://google.com/search?q=tiếng việt", "việt"},
		{"tìm kiếm trên linux   ", "linux"},
		{"một từ", "từ"},
	}

	for _, tc := range testCases {
		actual := getLastWordFromSentence(tc.sentence)
		if actual != tc.expected {
			t.Errorf("getLastWordFromSentence(%q) = %q; want %q", tc.sentence, actual, tc.expected)
		}
	}
}

// TestPreeditCommitBeforeHide verifies BUG-01 logic in preedit mode
func TestPreeditCommitBeforeHide(t *testing.T) {
	imDefs := bamboo.GetInputMethodDefinitions()
	telexMethod := bamboo.ParseInputMethod(imDefs, "Telex")
	preeditor := bamboo.NewEngine(telexMethod, bamboo.EstdFlags)

	cfg := config.DefaultCfg()
	cfg.DefaultInputMode = config.PreeditIM
	base := NewFakeEngine()
	engine := NewIbusBambooEngine("TestEngine", &cfg, base, preeditor)

	// Simulate gõ: "t", "i", "e", "e", "n", "g", "s"
	keys := []rune{'t', 'i', 'e', 'e', 'n', 'g', 's'}
	for _, k := range keys {
		processed, err := engine.ProcessKeyEvent(uint32(k), uint32(k), 0)
		if err != nil {
			t.Fatalf("Unexpected error on key '%c': %v", k, err)
		}
		if !processed {
			t.Fatalf("Key '%c' should be processed by engine", k)
		}
	}

	// Verify that current preedit text has been constructed
	preeditText := engine.getPreeditString()
	if preeditText != "tiếng" {
		t.Errorf("Expected preeditText 'tiếng', got '%s'", preeditText)
	}

	// Press space (word break)
	spaceProcessed, err := engine.ProcessKeyEvent(uint32(' '), uint32(' '), 0)
	if err != nil {
		t.Fatalf("Error on space key: %v", err)
	}
	if !spaceProcessed {
		t.Fatalf("Space should be processed to commit the preedited word")
	}

	// After space, preedit buffer should be reset
	if engine.getPreeditString() != "" {
		t.Errorf("Preedit buffer should be empty after word break, got '%s'", engine.getPreeditString())
	}
}

// TestSetSurroundingTextSelectionAndBoundaries verifies BUG-04 and BUG-05 fixes:
// selection-awareness and safe word boundary isolation during SetSurroundingText.
func TestSetSurroundingTextSelectionAndBoundaries(t *testing.T) {
	imDefs := bamboo.GetInputMethodDefinitions()
	telexMethod := bamboo.ParseInputMethod(imDefs, "Telex")
	preeditor := bamboo.NewEngine(telexMethod, bamboo.EstdFlags)

	cfg := config.DefaultCfg()
	cfg.DefaultInputMode = config.SurroundingTextIM
	base := NewFakeEngine()
	engine := NewIbusBambooEngine("TestEngine", &cfg, base, preeditor)

	// Case 1: Active selection in browser address bar (e.g. typing "viet" with autocomplete suggestion "nam")
	// text: "vietnam", cursorPos: 7, anchorPos: 4 (selection is from index 4 to 7)
	engine.isSurroundingTextReady = true
	variant1 := dbus.MakeVariant([]interface{}{"", nil, "https://google.com/search?q=tieng"})
	err := engine.SetSurroundingText(variant1, 37, 37)
	if err != nil {
		t.Fatalf("Unexpected error in SetSurroundingText: %v", err)
	}
	if engine.currentWordNearCursor != "tieng" {
		t.Errorf("Expected current word 'tieng', got %q", engine.currentWordNearCursor)
	}

	// Case 2: Address bar with inline autocomplete selected
	// Full text: "https://vietnamnet.vn", user typed up to "viet", remaining is selected
	// anchorPos = 12, cursorPos = 21
	engine.isSurroundingTextReady = true
	variant2 := dbus.MakeVariant([]interface{}{"", nil, "https://vietnamnet.vn"})
	err = engine.SetSurroundingText(variant2, 21, 12)
	if err != nil {
		t.Fatalf("Unexpected error in SetSurroundingText with selection: %v", err)
	}
	if engine.currentWordNearCursor != "viet" {
		t.Errorf("Expected current word before selection 'viet', got %q", engine.currentWordNearCursor)
	}

	// Case 3: Out-of-bounds cursor safety
	engine.isSurroundingTextReady = true
	variant3 := dbus.MakeVariant([]interface{}{"", nil, "short"})
	err = engine.SetSurroundingText(variant3, 9999, 9999)
	if err != nil {
		t.Fatalf("Out-of-bounds cursor should not return error: %v", err)
	}
	if engine.currentWordNearCursor != "short" {
		t.Errorf("Expected clamped current word 'short', got %q", engine.currentWordNearCursor)
	}
}

// TestFastToneModifications verifies rapid switching of tone marks on a single syllable
func TestFastToneModifications(t *testing.T) {
	imDefs := bamboo.GetInputMethodDefinitions()
	telexMethod := bamboo.ParseInputMethod(imDefs, "Telex")

	engine := bamboo.NewEngine(telexMethod, bamboo.EstdFlags)
	// Type "toan"
	for _, r := range "toan" {
		engine.ProcessKey(r, bamboo.VietnameseMode)
	}
	if engine.GetProcessedString(bamboo.VietnameseMode) != "toan" {
		t.Fatalf("Expected 'toan', got '%s'", engine.GetProcessedString(bamboo.VietnameseMode))
	}

	// Change to "toán" with 's'
	engine.ProcessKey('s', bamboo.VietnameseMode)
	if engine.GetProcessedString(bamboo.VietnameseMode) != "toán" {
		t.Errorf("Expected 'toán', got '%s'", engine.GetProcessedString(bamboo.VietnameseMode))
	}

	// Change to "toàn" with 'f'
	engine.ProcessKey('f', bamboo.VietnameseMode)
	if engine.GetProcessedString(bamboo.VietnameseMode) != "toàn" {
		t.Errorf("Expected 'toàn', got '%s'", engine.GetProcessedString(bamboo.VietnameseMode))
	}

	// Change to "toản" with 'r'
	engine.ProcessKey('r', bamboo.VietnameseMode)
	if engine.GetProcessedString(bamboo.VietnameseMode) != "toản" {
		t.Errorf("Expected 'toản', got '%s'", engine.GetProcessedString(bamboo.VietnameseMode))
	}

	// Change to "toãn" with 'x'
	engine.ProcessKey('x', bamboo.VietnameseMode)
	if engine.GetProcessedString(bamboo.VietnameseMode) != "toãn" {
		t.Errorf("Expected 'toãn', got '%s'", engine.GetProcessedString(bamboo.VietnameseMode))
	}

	// Change to "toạn" with 'j'
	engine.ProcessKey('j', bamboo.VietnameseMode)
	if engine.GetProcessedString(bamboo.VietnameseMode) != "toạn" {
		t.Errorf("Expected 'toạn', got '%s'", engine.GetProcessedString(bamboo.VietnameseMode))
	}
}

// TestResolvePresetInputMode verifies Phase 4: automatic preset recognition for Electron, browsers, editors
func TestResolvePresetInputMode(t *testing.T) {
	testCases := []struct {
		wmClass      string
		expectedMode int
		expectedOk   bool
	}{
		{"google-chrome:Google-chrome", config.SurroundingTextIM, true},
		{"google-chrome", config.SurroundingTextIM, true},
		{"chromium", config.SurroundingTextIM, true},
		{"brave-browser", config.SurroundingTextIM, true},
		{"microsoft-edge", config.SurroundingTextIM, true},
		{"firefox:Navigator", config.SurroundingTextIM, true},
		{"code:Code", config.SurroundingTextIM, true},
		{"cursor:Cursor", config.SurroundingTextIM, true},
		{"vscodium", config.SurroundingTextIM, true},
		{"zalo:Zalo", config.SurroundingTextIM, true},
		{"zalopc", config.SurroundingTextIM, true},
		{"slack:Slack", config.SurroundingTextIM, true},
		{"discord", config.SurroundingTextIM, true},
		{"teams", config.SurroundingTextIM, true},
		{"soffice.bin", config.SurroundingTextIM, true},
		{"libreoffice", config.SurroundingTextIM, true},
		{"gnome-terminal-server:Gnome-terminal", config.ForwardAsCommitIM, true},
		{"kitty", config.ForwardAsCommitIM, true},
		{"alacritty", config.ForwardAsCommitIM, true},
		{"unknown-custom-app-12345", 0, false},
		{"", 0, false},
	}

	for _, tc := range testCases {
		t.Run("Preset_"+tc.wmClass, func(t *testing.T) {
			mode, ok := config.ResolvePresetInputMode(tc.wmClass)
			if ok != tc.expectedOk {
				t.Errorf("ResolvePresetInputMode(%q) ok = %v; want %v", tc.wmClass, ok, tc.expectedOk)
			}
			if ok && mode != tc.expectedMode {
				t.Errorf("ResolvePresetInputMode(%q) mode = %d; want %d", tc.wmClass, mode, tc.expectedMode)
			}
		})
	}
}

// TestGetInputModePriority verifies priority: User manual mapping > Built-in preset > DefaultInputMode
func TestGetInputModePriority(t *testing.T) {
	imDefs := bamboo.GetInputMethodDefinitions()
	telexMethod := bamboo.ParseInputMethod(imDefs, "Telex")
	preeditor := bamboo.NewEngine(telexMethod, bamboo.EstdFlags)

	cfg := config.DefaultCfg()
	cfg.DefaultInputMode = config.PreeditIM
	// User explicitly overrides VSCode to ForwardAsCommitIM and Chrome to UsIM
	cfg.InputModeMapping = map[string]int{
		"code:Code":                  config.ForwardAsCommitIM,
		"google-chrome:Google-chrome": config.UsIM,
	}

	base := NewFakeEngine()
	engine := NewIbusBambooEngine("TestEngine", &cfg, base, preeditor)

	// 1. User manual override for VSCode
	engine.wmClasses = "code:Code"
	if engine.getInputMode() != config.ForwardAsCommitIM {
		t.Errorf("Expected user override mode %d for VSCode, got %d", config.ForwardAsCommitIM, engine.getInputMode())
	}

	// 2. User manual override for Chrome
	engine.wmClasses = "google-chrome:Google-chrome"
	if engine.getInputMode() != config.UsIM {
		t.Errorf("Expected user override mode %d for Chrome, got %d", config.UsIM, engine.getInputMode())
	}

	// 3. Built-in preset for Zalo (not in user mapping, should resolve to SurroundingTextIM)
	engine.wmClasses = "zalo:Zalo"
	if engine.getInputMode() != config.SurroundingTextIM {
		t.Errorf("Expected preset mode %d for Zalo, got %d", config.SurroundingTextIM, engine.getInputMode())
	}

	// 4. Built-in preset for Terminal (should resolve to ForwardAsCommitIM)
	engine.wmClasses = "kitty"
	if engine.getInputMode() != config.ForwardAsCommitIM {
		t.Errorf("Expected preset mode %d for Kitty, got %d", config.ForwardAsCommitIM, engine.getInputMode())
	}

	// 5. Unknown app (should fallback to DefaultInputMode = PreeditIM)
	engine.wmClasses = "random-unknown-app-xyz"
	if engine.getInputMode() != config.PreeditIM {
		t.Errorf("Expected fallback mode %d for unknown app, got %d", config.PreeditIM, engine.getInputMode())
	}
}

