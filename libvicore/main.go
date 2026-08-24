package main

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"sync"
	"unsafe"

	"github.com/BambooEngine/bamboo-core"
)

type EngineInstance struct {
	mu            sync.Mutex
	engine        bamboo.IEngine
	preeditString string
	rawKeys       []rune
}

var (
	instancesMu sync.RWMutex
	instances   = make(map[uintptr]*EngineInstance)
	nextID      uintptr = 1
)

//export vi_engine_new
func vi_engine_new(inputMethodName *C.char, flags C.uint) C.uintptr_t {
	imName := "Telex"
	if inputMethodName != nil {
		imName = C.GoString(inputMethodName)
	}

	imDefs := bamboo.GetInputMethodDefinitions()
	im := bamboo.ParseInputMethod(imDefs, imName)
	coreEngine := bamboo.NewEngine(im, uint(flags))

	instancesMu.Lock()
	defer instancesMu.Unlock()

	id := nextID
	nextID++
	instances[id] = &EngineInstance{
		engine: coreEngine,
	}

	return C.uintptr_t(id)
}

//export vi_engine_free
func vi_engine_free(handle C.uintptr_t) {
	instancesMu.Lock()
	defer instancesMu.Unlock()

	delete(instances, uintptr(handle))
}

//export vi_engine_reset
func vi_engine_reset(handle C.uintptr_t) {
	instancesMu.RLock()
	inst, ok := instances[uintptr(handle)]
	instancesMu.RUnlock()

	if !ok {
		return
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()

	inst.engine.Reset()
	inst.preeditString = ""
	inst.rawKeys = nil
}

//export vi_engine_process_key
func vi_engine_process_key(
	handle C.uintptr_t,
	keyVal C.uint,
	keyCode C.uint,
	state C.uint,
	outCommit *C.char,
	maxCommitLen C.int,
	outPreedit *C.char,
	maxPreeditLen C.int,
	outBsCount *C.int,
) C.int {
	instancesMu.RLock()
	inst, ok := instances[uintptr(handle)]
	instancesMu.RUnlock()

	if !ok {
		return 0
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()

	keyRune := rune(keyVal)

	// Modifier filtering: ignore shortcuts like Ctrl, Alt, Super
	const (
		ShiftMask   = 1 << 0
		LockMask    = 1 << 1
		ControlMask = 1 << 2
		Mod1Mask    = 1 << 3 // Alt
		Mod4Mask    = 1 << 6 // Super
	)

	if uint(state)&(ControlMask|Mod1Mask|Mod4Mask) != 0 {
		return 0
	}

	// Backspace handling
	const BackSpaceKeyVal = 0xff08
	if keyVal == BackSpaceKeyVal {
		if len(inst.rawKeys) > 0 {
			inst.engine.RemoveLastChar(true)
			inst.rawKeys = inst.rawKeys[:len(inst.rawKeys)-1]
			inst.preeditString = inst.engine.GetProcessedString(bamboo.VietnameseMode)
			copyStringToC(inst.preeditString, outPreedit, int(maxPreeditLen))
			return 1
		}
		return 0
	}

	// Return key handling
	const ReturnKeyVal = 0xff0d
	if keyVal == ReturnKeyVal {
		if len(inst.preeditString) > 0 {
			copyStringToC(inst.preeditString, outCommit, int(maxCommitLen))
			inst.engine.Reset()
			inst.preeditString = ""
			inst.rawKeys = nil
			if outBsCount != nil {
				*outBsCount = 0
			}
			return 1
		}
		return 0
	}

	// Process printable key with Vietnamese engine (handles both Telex keys and VNI numeric keys)
	if inst.engine.CanProcessKey(keyRune) {
		if uint(state)&LockMask != 0 {
			// Capslock
			if keyRune >= 'a' && keyRune <= 'z' {
				keyRune = keyRune - 32
			}
		}

		oldLen := len([]rune(inst.preeditString))
		inst.engine.ProcessKey(keyRune, bamboo.VietnameseMode)
		inst.rawKeys = append(inst.rawKeys, keyRune)
		newPreedit := inst.engine.GetProcessedString(bamboo.VietnameseMode)

		if outBsCount != nil && oldLen > 0 {
			bs := oldLen
			*outBsCount = C.int(bs)
		}

		inst.preeditString = newPreedit
		copyStringToC(inst.preeditString, outPreedit, int(maxPreeditLen))
		return 1
	}

	// Word break symbol handling (space, punctuation marks)
	if bamboo.IsWordBreakSymbol(keyRune) {
		if len(inst.preeditString) > 0 {
			commitText := inst.preeditString + string(keyRune)
			copyStringToC(commitText, outCommit, int(maxCommitLen))
			inst.engine.Reset()
			inst.preeditString = ""
			inst.rawKeys = nil
			if outBsCount != nil {
				*outBsCount = 0
			}
			return 1
		}
		return 0
	}

	return 0
}

//export vi_engine_set_surrounding_text
func vi_engine_set_surrounding_text(handle C.uintptr_t, text *C.char, cursorPos C.uint, anchorPos C.uint) {
	instancesMu.RLock()
	inst, ok := instances[uintptr(handle)]
	instancesMu.RUnlock()

	if !ok || text == nil {
		return
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()

	goStr := C.GoString(text)
	s := []rune(goStr)

	effectivePos := cursorPos
	if anchorPos < cursorPos {
		effectivePos = anchorPos
	}

	if int(effectivePos) > len(s) {
		effectivePos = C.uint(len(s))
	}

	cs := s[:effectivePos]
	startIdx := 0
	for i := len(cs) - 1; i >= 0; i-- {
		if bamboo.IsWordBreakSymbol(cs[i]) {
			startIdx = i + 1
			break
		}
	}

	wordRunes := cs[startIdx:]
	inst.engine.Reset()
	inst.rawKeys = nil
	for i := len(wordRunes) - 1; i >= 0; i-- {
		inst.engine.ProcessKey(wordRunes[i], bamboo.EnglishMode|bamboo.InReverseOrder)
	}
	inst.preeditString = string(wordRunes)
}

//export vi_engine_version
func vi_engine_version() *C.char {
	return C.CString("1.0.0-bamboo-viet")
}

func copyStringToC(src string, dst *C.char, maxLen int) {
	if dst == nil || maxLen <= 0 {
		return
	}
	srcBytes := []byte(src)
	n := len(srcBytes)
	if n >= maxLen {
		n = maxLen - 1
	}
	if n > 0 {
		C.memcpy(unsafe.Pointer(dst), unsafe.Pointer(&srcBytes[0]), C.size_t(n))
	}
	// Null terminate
	ptr := (*[1 << 20]byte)(unsafe.Pointer(dst))
	ptr[n] = 0
}

func main() {}
