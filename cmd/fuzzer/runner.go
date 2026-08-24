package main

import (
	"strings"

	"github.com/BambooEngine/bamboo-core"
)

// EngineLevel represents which level of the stack was tested
type EngineLevel string

const (
	LevelBambooCore EngineLevel = "bamboo-core (Lõi FSM)"
	LevelLibViCore   EngineLevel = "libvicore (C ABI)"
	LevelIBusEngine  EngineLevel = "IBusBambooEngine (Full Pipeline)"
)

// ExecutionResult holds the outcome of running a scenario
type ExecutionResult struct {
	Scenario      TypingScenario
	Level         EngineLevel
	InputMethod   string
	ActualOutput  string
	Passed        bool
	ErrorMessage  string
	ErrorType     ErrorType
	CommittedText string
	PreeditText   string
}

// FuzzRunner executes test scenarios against different engine layers
type FuzzRunner struct {
	imDefs      map[string]bamboo.InputMethodDefinition
	telexMethod bamboo.InputMethod
	vniMethod   bamboo.InputMethod
}

func NewFuzzRunner() *FuzzRunner {
	defs := bamboo.GetInputMethodDefinitions()
	return &FuzzRunner{
		imDefs:      defs,
		telexMethod: bamboo.ParseInputMethod(defs, "Telex"),
		vniMethod:   bamboo.ParseInputMethod(defs, "VNI"),
	}
}

// RunBambooCore runs a scenario directly against bamboo-core
func (r *FuzzRunner) RunBambooCore(sc TypingScenario, inputMethod string) ExecutionResult {
	im := r.telexMethod
	if strings.EqualFold(inputMethod, "vni") {
		im = r.vniMethod
	}

	engine := bamboo.NewEngine(im, bamboo.EstdFlags)
	var committed strings.Builder

	for _, ks := range sc.Keystrokes {
		if ks.IsBS {
			engine.RemoveLastChar(true)
		} else if bamboo.IsWordBreakSymbol(ks.Char) {
			word := engine.GetProcessedString(bamboo.VietnameseMode)
			committed.WriteString(word)
			committed.WriteRune(ks.Char)
			engine.Reset()
		} else if engine.CanProcessKey(ks.Char) {
			engine.ProcessKey(ks.Char, bamboo.VietnameseMode)
		} else {
			committed.WriteRune(ks.Char)
		}
	}

	var finalWord string
	if !engine.IsValid(false) && bamboo.HasAnyVietnameseRune(engine.GetProcessedString(bamboo.VietnameseMode)) {
		finalWord = engine.GetProcessedString(bamboo.EnglishMode)
	} else {
		finalWord = engine.GetProcessedString(bamboo.VietnameseMode)
	}
	finalOutput := committed.String() + finalWord

	passed := (finalOutput == sc.TargetText || isTonePlacementEquivalent(sc.TargetText, finalOutput))
	errType := None
	var errMsg string

	if !passed {
		errType, errMsg = ClassifyError(sc.TargetText, finalOutput, sc.InputSequence)
	}

	return ExecutionResult{
		Scenario:      sc,
		Level:         LevelBambooCore,
		InputMethod:   inputMethod,
		ActualOutput:  finalOutput,
		Passed:        passed,
		ErrorType:     errType,
		ErrorMessage:  errMsg,
		CommittedText: committed.String(),
		PreeditText:   finalWord,
	}
}

// RunSimulatedIBusEngine simulates full IBus pipeline behavior with buffer, word break, backspacing
func (r *FuzzRunner) RunSimulatedIBusEngine(sc TypingScenario, inputMethod string, mode int) ExecutionResult {
	im := r.telexMethod
	if strings.EqualFold(inputMethod, "vni") {
		im = r.vniMethod
	}

	engine := bamboo.NewEngine(im, bamboo.EstdFlags)
	var committed strings.Builder
	var lastWord string

	for _, ks := range sc.Keystrokes {
		if ks.IsBS {
			if len(lastWord) > 0 {
				engine.RemoveLastChar(true)
				lastWord = engine.GetProcessedString(bamboo.VietnameseMode)
			}
		} else if bamboo.IsWordBreakSymbol(ks.Char) {
			word := engine.GetProcessedString(bamboo.VietnameseMode)
			committed.WriteString(word)
			committed.WriteRune(ks.Char)
			engine.Reset()
			lastWord = ""
		} else if engine.CanProcessKey(ks.Char) {
			engine.ProcessKey(ks.Char, bamboo.VietnameseMode)
			lastWord = engine.GetProcessedString(bamboo.VietnameseMode)
		} else {
			committed.WriteRune(ks.Char)
		}
	}

	finalOutput := committed.String() + lastWord
	passed := (finalOutput == sc.TargetText || isTonePlacementEquivalent(sc.TargetText, finalOutput))
	errType := None
	var errMsg string

	if !passed {
		errType, errMsg = ClassifyError(sc.TargetText, finalOutput, sc.InputSequence)
	}

	return ExecutionResult{
		Scenario:      sc,
		Level:         LevelIBusEngine,
		InputMethod:   inputMethod,
		ActualOutput:  finalOutput,
		Passed:        passed,
		ErrorType:     errType,
		ErrorMessage:  errMsg,
		CommittedText: committed.String(),
		PreeditText:   lastWord,
	}
}
