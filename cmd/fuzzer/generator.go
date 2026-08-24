package main

import (
	"math/rand"
	"strings"
	"time"
	"unicode"

	"github.com/BambooEngine/bamboo-core"
)

// TypingStyle represents different human typing behaviors
type TypingStyle int

const (
	StyleStandardLateTone TypingStyle = iota // Bỏ dấu cuối từ: "tieengs", "hoaf", "ddeer"
	StyleStandardMidTone                    // Bỏ dấu ngay sau nguyên âm: "tiseng", "hoas"
	StyleRapidToneChange                    // Đổi dấu liên tục: "toans" -> "toanf" -> "toanr"
	StyleTypoWithBackspace                  // Gõ sai rồi bấm Backspace sửa
	StyleFastOvertyping                     // Gõ lặp phím: "ddeer", "tooi", "vuwaf"
	StyleWordBoundary                       // Gõ nối từ không ngắt quãng
)

// Keystroke represents a single key event simulation
type Keystroke struct {
	KeyVal   uint32
	KeyCode  uint32
	State    uint32
	Char     rune
	IsBS     bool
	DelayMs  int
}

// TypingScenario represents a test scenario generated for a target word/sentence
type TypingScenario struct {
	TargetText    string
	InputSequence string
	Style         TypingStyle
	Keystrokes    []Keystroke
	Description   string
}

// Map of Vietnamese character decompositions
var vnCharToTelex = map[rune]struct {
	base rune
	mod  string
	tone string
}{
	'á': {'a', "", "s"}, 'à': {'a', "", "f"}, 'ả': {'a', "", "r"}, 'ã': {'a', "", "x"}, 'ạ': {'a', "", "j"},
	'ắ': {'a', "w", "s"}, 'ằ': {'a', "w", "f"}, 'ẳ': {'a', "w", "r"}, 'ẵ': {'a', "w", "x"}, 'ặ': {'a', "w", "j"}, 'ă': {'a', "w", ""},
	'ấ': {'a', "a", "s"}, 'ầ': {'a', "a", "f"}, 'ẩ': {'a', "a", "r"}, 'ẫ': {'a', "a", "x"}, 'ậ': {'a', "a", "j"}, 'â': {'a', "a", ""},
	'é': {'e', "", "s"}, 'è': {'e', "", "f"}, 'ẻ': {'e', "", "r"}, 'ẽ': {'e', "", "x"}, 'ẹ': {'e', "", "j"},
	'ế': {'e', "e", "s"}, 'ề': {'e', "e", "f"}, 'ể': {'e', "e", "r"}, 'ễ': {'e', "e", "x"}, 'ệ': {'e', "e", "j"}, 'ê': {'e', "e", ""},
	'í': {'i', "", "s"}, 'ì': {'i', "", "f"}, 'ỉ': {'i', "", "r"}, 'ĩ': {'i', "", "x"}, 'ị': {'i', "", "j"},
	'ó': {'o', "", "s"}, 'ò': {'o', "", "f"}, 'ỏ': {'o', "", "r"}, 'õ': {'o', "", "x"}, 'ọ': {'o', "", "j"},
	'ố': {'o', "o", "s"}, 'ồ': {'o', "o", "f"}, 'ổ': {'o', "o", "r"}, 'ỗ': {'o', "o", "x"}, 'ộ': {'o', "o", "j"}, 'ô': {'o', "o", ""},
	'ớ': {'o', "w", "s"}, 'ờ': {'o', "w", "f"}, 'ở': {'o', "w", "r"}, 'ỡ': {'o', "w", "x"}, 'ợ': {'o', "w", "j"}, 'ơ': {'o', "w", ""},
	'ú': {'u', "", "s"}, 'ù': {'u', "", "f"}, 'ủ': {'u', "", "r"}, 'ũ': {'u', "", "x"}, 'ụ': {'u', "", "j"},
	'ứ': {'u', "w", "s"}, 'ừ': {'u', "w", "f"}, 'ử': {'u', "w", "r"}, 'ữ': {'u', "w", "x"}, 'ự': {'u', "w", "j"}, 'ư': {'u', "w", ""},
	'ý': {'y', "", "s"}, 'ỳ': {'y', "", "f"}, 'ỷ': {'y', "", "r"}, 'ỹ': {'y', "", "x"}, 'ỵ': {'y', "", "j"},
	'đ': {'d', "d", ""}, 'Đ': {'d', "d", ""},
}

// DecomposeWordToTelex turns any Vietnamese word into valid human Telex typing sequences
func DecomposeWordToTelex(word string) []string {
	cleanWord := strings.Trim(word, " ,.!?:;\"'()[]{}")
	if cleanWord == "" {
		return nil
	}

	var rawBase strings.Builder
	var toneKey string
	var modKeys strings.Builder

	runes := []rune(cleanWord)
	for _, r := range runes {
		lowerR := unicode.ToLower(r)
		if decomp, ok := vnCharToTelex[lowerR]; ok {
			rawBase.WriteRune(decomp.base)
			if decomp.mod != "" {
				modKeys.WriteString(decomp.mod)
			}
			if decomp.tone != "" {
				toneKey = decomp.tone
			}
		} else {
			rawBase.WriteRune(lowerR)
		}
	}

	// Build common variations:
	// 1. Late tone: spell base with modifier + tone at the end
	var lateSeq strings.Builder
	for _, r := range runes {
		lowerR := unicode.ToLower(r)
		if decomp, ok := vnCharToTelex[lowerR]; ok {
			rawBase.WriteRune(decomp.base)
			if lowerR == 'đ' {
				lateSeq.WriteString("dd")
			} else if decomp.mod == "w" {
				lateSeq.WriteRune(decomp.base)
				lateSeq.WriteString("w")
			} else if decomp.mod == "a" || decomp.mod == "e" || decomp.mod == "o" {
				lateSeq.WriteRune(decomp.base)
				lateSeq.WriteString(decomp.mod)
			} else {
				lateSeq.WriteRune(decomp.base)
			}
		} else {
			lateSeq.WriteRune(lowerR)
		}
	}
	if toneKey != "" {
		lateSeq.WriteString(toneKey)
	}

	// Normalize ươ in lateSeq to standard uow
	seq := lateSeq.String()
	seq = strings.ReplaceAll(seq, "uwow", "uow")
	seq = strings.ReplaceAll(seq, "uww", "uow")

	res := []string{seq}
	return res
}

// ScenarioGenerator builds realistic test scenarios
type ScenarioGenerator struct {
	rnd *rand.Rand
}

func NewScenarioGenerator() *ScenarioGenerator {
	return &ScenarioGenerator{
		rnd: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateForWord generates multiple realistic typing scenarios for a single word
func (g *ScenarioGenerator) GenerateForWord(word string) []TypingScenario {
	cleanWord := strings.Trim(word, " ,.!?:;\"'()[]{}")
	if cleanWord == "" {
		return nil
	}

	var scenarios []TypingScenario
	variants := DecomposeWordToTelex(cleanWord)
	if len(variants) == 0 {
		variants = []string{cleanWord}
	}

	for _, v := range variants {
		// Scenario 1: Standard keystroke sequence
		scenarios = append(scenarios, TypingScenario{
			TargetText:    cleanWord,
			InputSequence: v,
			Style:         StyleStandardLateTone,
			Keystrokes:    g.stringToKeystrokes(v),
			Description:   "Standard typing: " + v,
		})

		// Scenario 2: Rapid tone modifications (e.g. gõ nhầm dấu rồi sửa lại)
		toneSwitches := g.generateToneSwitches(v, cleanWord)
		for _, ts := range toneSwitches {
			scenarios = append(scenarios, TypingScenario{
				TargetText:    cleanWord,
				InputSequence: ts,
				Style:         StyleRapidToneChange,
				Keystrokes:    g.stringToKeystrokes(ts),
				Description:   "Rapid tone switch: " + ts,
			})
		}

		// Scenario 3: Typos with Backspace mid-word
		typoSeq := g.generateTypoWithBackspace(v)
		if len(typoSeq) > 0 {
			scenarios = append(scenarios, TypingScenario{
				TargetText:    cleanWord,
				InputSequence: g.keystrokesToString(typoSeq),
				Style:         StyleTypoWithBackspace,
				Keystrokes:    typoSeq,
				Description:   "Typo with backspace mid-word",
			})
		}
	}

	return scenarios
}

// GenerateSentenceScenarios breaks a sentence into words and generates mixed typing flows
func (g *ScenarioGenerator) GenerateSentenceScenarios(sentence string) TypingScenario {
	words := strings.Fields(sentence)
	var allKeystrokes []Keystroke
	var fullInputSeq strings.Builder

	for i, rawWord := range words {
		// Check trailing punctuation
		trailingPunct := ""
		w := rawWord
		if strings.HasSuffix(rawWord, ",") || strings.HasSuffix(rawWord, ".") || strings.HasSuffix(rawWord, "!") || strings.HasSuffix(rawWord, "?") {
			trailingPunct = string(rawWord[len(rawWord)-1])
			w = rawWord[:len(rawWord)-1]
		}

		scenarios := g.GenerateForWord(w)
		if len(scenarios) == 0 {
			continue
		}
		chosen := scenarios[0]

		allKeystrokes = append(allKeystrokes, chosen.Keystrokes...)
		fullInputSeq.WriteString(chosen.InputSequence)

		if trailingPunct != "" {
			pChar := rune(trailingPunct[0])
			allKeystrokes = append(allKeystrokes, Keystroke{
				KeyVal:  uint32(pChar),
				KeyCode: uint32(pChar),
				Char:    pChar,
			})
			fullInputSeq.WriteString(trailingPunct)
		}

		// Append space between words
		if i < len(words)-1 {
			allKeystrokes = append(allKeystrokes, Keystroke{
				KeyVal:  uint32(' '),
				KeyCode: uint32(' '),
				Char:    ' ',
			})
			fullInputSeq.WriteRune(' ')
		}
	}

	return TypingScenario{
		TargetText:    sentence,
		InputSequence: fullInputSeq.String(),
		Style:         StyleWordBoundary,
		Keystrokes:    allKeystrokes,
		Description:   "Full sentence: " + sentence,
	}
}

func (g *ScenarioGenerator) stringToKeystrokes(s string) []Keystroke {
	var ks []Keystroke
	for _, r := range s {
		var state uint32 = 0
		if unicode.IsUpper(r) {
			state = 1 // Shift
		}
		ks = append(ks, Keystroke{
			KeyVal:  uint32(r),
			KeyCode: uint32(r),
			State:   state,
			Char:    r,
		})
	}
	return ks
}

func (g *ScenarioGenerator) keystrokesToString(ks []Keystroke) string {
	var sb strings.Builder
	for _, k := range ks {
		if k.IsBS {
			sb.WriteString("<BS>")
		} else {
			sb.WriteRune(k.Char)
		}
	}
	return sb.String()
}

func (g *ScenarioGenerator) generateToneSwitches(baseSeq string, targetWord string) []string {
	// Only generate tone switches if the target word actually has Vietnamese tone mark
	if !bamboo.HasAnyVietnameseRune(targetWord) {
		return nil
	}
	var results []string
	if len(baseSeq) > 1 {
		lastRune := rune(baseSeq[len(baseSeq)-1])
		stem := baseSeq[:len(baseSeq)-1]
		isStopped := strings.HasSuffix(stem, "c") || strings.HasSuffix(stem, "p") || strings.HasSuffix(stem, "t") || strings.HasSuffix(stem, "ch")

		if lastRune == 's' || lastRune == 'f' || lastRune == 'r' || lastRune == 'x' || lastRune == 'j' {
			var otherTone rune = 's'
			if lastRune == 's' {
				if isStopped {
					otherTone = 'j'
				} else {
					otherTone = 'f'
				}
			} else if lastRune == 'j' {
				otherTone = 's'
			} else if isStopped {
				otherTone = 'j'
			} else {
				otherTone = 's'
			}
			results = append(results, stem+string(otherTone)+string(lastRune))
		}
	}
	return results
}

func (g *ScenarioGenerator) generateTypoWithBackspace(baseSeq string) []Keystroke {
	if len(baseSeq) < 3 {
		return nil
	}

	var ks []Keystroke
	runes := []rune(baseSeq)
	splitPoint := len(runes) / 2

	// Type first half
	for i := 0; i < splitPoint; i++ {
		ks = append(ks, Keystroke{KeyVal: uint32(runes[i]), KeyCode: uint32(runes[i]), Char: runes[i]})
	}

	// Insert a wrong typo character (e.g. 'm')
	typoChar := 'm'
	if runes[splitPoint] == 'm' {
		typoChar = 'k'
	}
	ks = append(ks, Keystroke{KeyVal: uint32(typoChar), KeyCode: uint32(typoChar), Char: typoChar})

	// Press BackSpace to delete the typo
	ks = append(ks, Keystroke{KeyVal: 0xff08, KeyCode: 0xff08, IsBS: true, Char: 0})

	// Type the remaining half
	for i := splitPoint; i < len(runes); i++ {
		ks = append(ks, Keystroke{KeyVal: uint32(runes[i]), KeyCode: uint32(runes[i]), Char: runes[i]})
	}

	return ks
}
