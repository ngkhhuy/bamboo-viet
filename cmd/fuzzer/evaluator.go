package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/BambooEngine/bamboo-core"
)

// ErrorType categorizes discrepancies found during fuzzing
type ErrorType string

const (
	None                       ErrorType = "NONE"
	ErrUntransformedRawKey     ErrorType = "ERR_UNTRANSFORMED_RAW_KEY"     // Phím thô không chuyển thành dấu (ddeer, vuwaf, thois)
	ErrSyllableDuplication     ErrorType = "ERR_SYLLABLE_DUPLICATION"     // Lặp từ / lặp vần (nguoiười)
	ErrSwallowedChar           ErrorType = "ERR_SWALLOWED_CHAR"           // Nuốt mất ký tự đầu/giữa (õ thay vì gõ)
	ErrTonePlacement           ErrorType = "ERR_TONE_PLACEMENT"           // Đặt sai vị trí dấu (hòa vs hoà)
	ErrBackspaceDesync         ErrorType = "ERR_BACKSPACE_DESYNC"         // Lệch số bước xóa lùi
	ErrOverRestoreFalsePositive ErrorType = "ERR_OVER_RESTORE"            // Khôi phục nhầm về tiếng Anh
	ErrUnknown                 ErrorType = "ERR_UNKNOWN"
)

func extractToneAndBase(word string) (string, string) {
	var base strings.Builder
	var foundTone string

	for _, r := range word {
		lowerR := unicode.ToLower(r)
		if decomp, ok := vnCharToTelex[lowerR]; ok {
			base.WriteRune(decomp.base)
			if decomp.mod != "" {
				base.WriteString(decomp.mod)
			}
			if decomp.tone != "" {
				foundTone = decomp.tone
			}
		} else {
			base.WriteRune(lowerR)
		}
	}
	return base.String(), foundTone
}

func isTonePlacementEquivalent(a, b string) bool {
	if a == b {
		return true
	}
	baseA, toneA := extractToneAndBase(a)
	baseB, toneB := extractToneAndBase(b)
	return baseA != "" && baseA == baseB && toneA == toneB
}

// ClassifyError inspects target vs actual output and determines the root cause
func ClassifyError(target, actual, inputSeq string) (ErrorType, string) {
	if target == actual || isTonePlacementEquivalent(target, actual) {
		return None, ""
	}

	targetRunes := []rune(target)
	actualRunes := []rune(actual)

	// 1. Check for Syllable Duplication (e.g. "nguoi" + "ười" -> "nguoiười")
	if len(actualRunes) > len(targetRunes) {
		if strings.Contains(actual, target) || strings.Contains(inputSeq, actual) {
			return ErrSyllableDuplication, fmt.Sprintf("Lặp từ/vần: mong đợi '%s', kết quả bị nhân đôi '%s'", target, actual)
		}
	}

	// 2. Check for Swallowed Characters (e.g. "õ" instead of "gõ")
	if len(actualRunes) < len(targetRunes) {
		if strings.HasSuffix(target, actual) {
			return ErrSwallowedChar, fmt.Sprintf("Nuốt ký tự đầu: mong đợi '%s', bị mất chữ thành '%s'", target, actual)
		}
	}

	// 3. Check for Untransformed Raw Keys (e.g. "ddeer" -> "ddeer", "vuwaf" -> "vuwaf", "thois" -> "thois")
	rawKeyMarkers := []string{"dd", "ee", "oo", "aa", "ow", "uw", "w", "s", "f", "r", "x", "j"}
	for _, marker := range rawKeyMarkers {
		if strings.Contains(actual, marker) && !strings.Contains(target, marker) {
			return ErrUntransformedRawKey, fmt.Sprintf("Phím thô không biến đổi dấu ('%s' còn sót trong '%s' thay vì '%s')", marker, actual, target)
		}
	}

	// 4. Check for Over-restore / English fallback
	if !bamboo.HasAnyVietnameseRune(actual) && bamboo.HasAnyVietnameseRune(target) {
		return ErrOverRestoreFalsePositive, fmt.Sprintf("Khôi phục tiếng Anh nhầm: mong đợi tiếng Việt '%s', bị trả về thô '%s'", target, actual)
	}

	// 5. Tone Placement difference (e.g. hòa vs hoà)
	if utf8.RuneCountInString(target) == utf8.RuneCountInString(actual) {
		return ErrTonePlacement, fmt.Sprintf("Khác biệt vị trí đặt dấu: mong đợi '%s', nhận được '%s'", target, actual)
	}

	return ErrUnknown, fmt.Sprintf("Sai lệch chưa xác định: mong đợi '%s', nhận được '%s' (chuỗi gõ: '%s')", target, actual, inputSeq)
}
