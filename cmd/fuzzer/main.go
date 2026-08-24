package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	modeFlag     = flag.String("mode", "telex", "Kiểu gõ: telex hoặc vni")
	countFlag    = flag.Int("count", 500, "Số lượng kịch bản gõ sinh ngẫu nhiên")
	sentenceFlag = flag.String("sentence", "", "Đoạn văn bản tùy chỉnh cần kiểm thử")
	stressFlag   = flag.Bool("stress", false, "Chạy chế độ stress test với toàn bộ biến thể phức tạp")
	reportFlag   = flag.String("report", "docs/fuzz-report.md", "Đường dẫn file báo cáo Markdown")
	genTestFlag  = flag.String("gen-test", "", "Đường dẫn sinh file Go test tái hiện lỗi (tùy chọn)")
	dictPathFlag = flag.String("dict", "data/vietnamese.cm.dict", "Đường dẫn file từ điển tiếng Việt")
)

func isPhonotacticallyValidVietnamese(w string) bool {
	if len(w) == 0 {
		return false
	}
	lower := strings.ToLower(w)
	// Filter out multi-syllable compounds or loanwords with non-Vietnamese initial clusters
	invalidPrefixes := []string{"uv", "ddt", "kr", "cr", "gr", "bl", "pl", "fl", "cl", "khl", "ddr", "eep", "ep", "êp", "đr", "đt"}
	for _, p := range invalidPrefixes {
		if strings.HasPrefix(lower, p) {
			return false
		}
	}

	// Filter out non-Vietnamese ending consonants (like -l, -k, -s, -d, -f, -r, -x, -j, -h (unless -ch, -nh))
	runes := []rune(lower)
	lastRune := runes[len(runes)-1]
	if lastRune == 'l' || lastRune == 'k' || lastRune == 's' || lastRune == 'd' || lastRune == 'f' || lastRune == 'r' || lastRune == 'x' || lastRune == 'j' ||
		(lastRune == 'g' && (len(runes) > 1 && runes[len(runes)-2] != 'n')) ||
		(lastRune == 'h' && (len(runes) > 1 && runes[len(runes)-2] != 'c' && runes[len(runes)-2] != 'n')) {
		return false
	}

	// Filter loanwords and invalid repeating vowels like 'ii', 'tête'
	if lower == "ii" || lower == "ìi" || lower == "tête" {
		return false
	}

	// In Vietnamese phonology: syllables ending in -c, -p, -t, -ch can only take SẮC or NẶNG
	isStopEnding := strings.HasSuffix(lower, "c") || strings.HasSuffix(lower, "p") || strings.HasSuffix(lower, "t") || strings.HasSuffix(lower, "ch")
	if isStopEnding {
		_, tone := extractToneAndBase(lower)
		if tone == "f" || tone == "r" || tone == "x" {
			return false
		}
	}

	// Filter non-standard interjections
	if lower == "eeng" || lower == "ềng" || lower == "ya" || lower == "ỵa" {
		return false
	}

	return true
}

func loadDictionaryWords(path string, maxWords int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() && len(words) < maxWords {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				w := parts[0]
				if isPhonotacticallyValidVietnamese(w) {
					words = append(words, w)
				}
			}
		}
	}
	return words, scanner.Err()
}

func main() {
	flag.Parse()

	fmt.Println("🎋 Bamboo Viet — Hệ Thống Fuzzing & Auto-Typing Thông Minh")
	fmt.Printf("Chế độ gõ : %s | Stress: %v | Báo cáo: %s\n", strings.ToUpper(*modeFlag), *stressFlag, *reportFlag)

	generator := NewScenarioGenerator()
	runner := NewFuzzRunner()
	reporter := NewReporter()

	// Case 1: Test specific sentence if provided
	if *sentenceFlag != "" {
		fmt.Printf("\n>>> Đang kiểm thử câu văn bản: %q\n", *sentenceFlag)
		scenario := generator.GenerateSentenceScenarios(*sentenceFlag)
		resCore := runner.RunBambooCore(scenario, *modeFlag)
		resIBus := runner.RunSimulatedIBusEngine(scenario, *modeFlag, 2)

		reporter.AddResult(resCore)
		reporter.AddResult(resIBus)

		// Test each word in the sentence individually with all variations
		words := strings.Fields(*sentenceFlag)
		for _, w := range words {
			scenarios := generator.GenerateForWord(w)
			for _, sc := range scenarios {
				r := runner.RunBambooCore(sc, *modeFlag)
				reporter.AddResult(r)
			}
		}
	} else {
		defaultWords := []string{
			"tiếng", "việt", "bộ", "gõ", "linux", "tôi", "muốn", "tự", "động", "để",
			"sửa", "lỗi", "tồn", "đọng", "vừa", "thói", "quen", "người", "hệ", "thống",
			"chính", "tả", "thuyền", "đường", "trường", "nguyễn", "phương", "quốc", "toàn",
		}
		targetWords := append([]string{}, defaultWords...)

		dictWords, err := loadDictionaryWords(*dictPathFlag, *countFlag)
		if err == nil {
			targetWords = append(targetWords, dictWords...)
		}

		limit := *countFlag
		if *stressFlag {
			limit = 5000
		}
		if len(targetWords) > limit {
			targetWords = targetWords[:limit]
		}

		fmt.Printf("Đang sinh kịch bản gõ cho %d từ vựng & ngữ cảnh tiếng Việt...\n", len(targetWords))

		for _, word := range targetWords {
			scenarios := generator.GenerateForWord(word)
			for _, sc := range scenarios {
				r := runner.RunBambooCore(sc, *modeFlag)
				reporter.AddResult(r)
			}
		}
	}

	// Print summary to terminal
	reporter.PrintConsoleSummary()

	// Generate Markdown report
	_ = os.MkdirAll(filepath.Dir(*reportFlag), 0755)
	if err := reporter.GenerateMarkdownReport(*reportFlag); err == nil {
		fmt.Printf("✓ Đã xuất báo cáo chi tiết tại: %s\n", *reportFlag)
	}

	// Generate Go test cases if there are errors
	if err := reporter.GenerateGoTestFile(*genTestFlag); err == nil {
		if _, err := os.Stat(*genTestFlag); err == nil {
			fmt.Printf("✓ Đã sinh file kiểm thử tái hiện lỗi: %s\n", *genTestFlag)
		}
	}
}
