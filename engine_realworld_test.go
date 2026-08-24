package main

import (
	"ibus-bamboo/config"
	"testing"

	"github.com/BambooEngine/bamboo-core"
)

func TestRealWorldTypingSequences(t *testing.T) {
	imDefs := bamboo.GetInputMethodDefinitions()
	telexMethod := bamboo.ParseInputMethod(imDefs, "Telex")

	testCases := []struct {
		name     string
		input    string
		expected string
		mode     int
	}{
		{"hopwj_preedit", "truwowngf hopwj ", "trường hợp ", config.PreeditIM},
		{"ddeer_preedit", "ddeer ", "để ", config.PreeditIM},
		{"vuwaf_preedit", "vuwaf ", "vừa ", config.PreeditIM},
		{"thois_preedit", "thois ", "thói ", config.PreeditIM},
		{"nguowif_preedit", "nguowif ", "người ", config.PreeditIM},
		{"duocwj_preedit", "duocwj ", "dược ", config.PreeditIM},
		{"nuocws_preedit", "nuocws ", "nước ", config.PreeditIM},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.DefaultCfg()
			cfg.DefaultInputMode = tc.mode
			cfg.IBflags = config.IBstdFlags
			pre := bamboo.NewEngine(telexMethod, cfg.Flags)
			base := NewFakeEngine()
			eng := NewIbusBambooEngine("TestEngine", &cfg, base, pre)

			for _, r := range tc.input {
				eng.ProcessKeyEvent(uint32(r), uint32(r), 0)
			}

			output := base.commitText

			if output != tc.expected {
				t.Errorf("Input %q yielded %q; want %q", tc.input, output, tc.expected)
			}
		})
	}
}
