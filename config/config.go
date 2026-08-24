package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/user"
	"strings"

	"github.com/BambooEngine/bamboo-core"
)

var DefaultAppPresets = map[string]int{
	// Browsers (Chromium / Gecko)
	"google-chrome":    SurroundingTextIM,
	"chrome":           SurroundingTextIM,
	"chromium":         SurroundingTextIM,
	"chromium-browser": SurroundingTextIM,
	"brave-browser":    SurroundingTextIM,
	"microsoft-edge":   SurroundingTextIM,
	"opera":            SurroundingTextIM,
	"vivaldi":          SurroundingTextIM,
	"firefox":          SurroundingTextIM,
	"navigator":        SurroundingTextIM,

	// Code Editors (Electron & Native)
	"code":         SurroundingTextIM,
	"vscodium":     SurroundingTextIM,
	"cursor":       SurroundingTextIM,
	"sublime_text": SurroundingTextIM,
	"atom":         SurroundingTextIM,

	// Chat Apps (Electron / Web / Desktop)
	"zalo":             SurroundingTextIM,
	"zalopc":           SurroundingTextIM,
	"slack":            SurroundingTextIM,
	"discord":          SurroundingTextIM,
	"teams":            SurroundingTextIM,
	"ms-teams":         SurroundingTextIM,
	"telegram-desktop": SurroundingTextIM,
	"skype":            SurroundingTextIM,

	// Office & Text Editors
	"soffice.bin":       SurroundingTextIM,
	"libreoffice":       SurroundingTextIM,
	"gedit":             SurroundingTextIM,
	"gnome-text-editor": SurroundingTextIM,

	// Terminals (Standalone Emulators)
	"gnome-terminal":        ForwardAsCommitIM,
	"gnome-terminal-server": ForwardAsCommitIM,
	"alacritty":             ForwardAsCommitIM,
	"kitty":                 ForwardAsCommitIM,
	"konsole":               ForwardAsCommitIM,
	"wezterm":               ForwardAsCommitIM,
	"wezterm-gui":           ForwardAsCommitIM,
	"xterm":                 ForwardAsCommitIM,
	"terminator":            ForwardAsCommitIM,
	"tilix":                 ForwardAsCommitIM,
	"foot":                  ForwardAsCommitIM,
	"guake":                 ForwardAsCommitIM,
	"tilda":                 ForwardAsCommitIM,
	"rxvt":                  ForwardAsCommitIM,
	"urxvt":                 ForwardAsCommitIM,
	"ghostty":               ForwardAsCommitIM,
	"tabby":                 ForwardAsCommitIM,
	"hyper":                 ForwardAsCommitIM,
}

func ResolvePresetInputMode(wmClass string) (int, bool) {
	if wmClass == "" {
		return 0, false
	}
	lowerWm := strings.ToLower(wmClass)

	// 1. Direct match
	if im, ok := DefaultAppPresets[lowerWm]; ok {
		return im, true
	}

	// 2. Component match (instance:class or reverse domain)
	parts := strings.FieldsFunc(lowerWm, func(r rune) bool {
		return r == ':' || r == '.' || r == ' '
	})
	for _, part := range parts {
		if im, ok := DefaultAppPresets[part]; ok {
			return im, true
		}
	}

	// 3. Substring match
	for key, im := range DefaultAppPresets {
		if strings.Contains(lowerWm, key) {
			return im, true
		}
	}

	return 0, false
}

const (
	configDir        = "%s/.config/ibus-%s"
	configFile       = "%s/ibus-%s.config.json"
	mactabFile       = "%s/ibus-%s.macro.text"
	sampleMactabFile = "data/macro.tpl.txt"
)

type Config struct {
	InputMethod            string
	InputMethodDefinitions map[string]bamboo.InputMethodDefinition
	OutputCharset          string
	Flags                  uint
	IBflags                uint
	Shortcuts              [10]uint32
	DefaultInputMode       int
	InputModeMapping       map[string]int
}

func GetConfigDir(ngName string) string {
	u, err := user.Current()
	if err == nil {
		return fmt.Sprintf(configDir, u.HomeDir, "bamboo")
	}
	return fmt.Sprintf(configDir, "~", "bamboo")
}

func GetMacroPath(engineName string) string {
	return fmt.Sprintf(mactabFile, GetConfigDir(engineName), engineName)
}

func GetConfigPath(engineName string) string {
	return fmt.Sprintf(configFile, GetConfigDir(engineName), engineName)
}

func DefaultCfg() Config {
	return Config{
		InputMethod:            "Telex",
		OutputCharset:          "Unicode",
		InputMethodDefinitions: bamboo.GetInputMethodDefinitions(),
		Flags:                  bamboo.EstdFlags,
		IBflags:                IBstdFlags,
		Shortcuts:              [10]uint32{1, 126, 0, 0, 0, 0, 0, 0, 5, 117},
		DefaultInputMode:       PreeditIM,
		InputModeMapping:       map[string]int{},
	}
}

func LoadConfig(engineName string) *Config {
	var c = DefaultCfg()
	if engineName == "bamboous" {
		c.DefaultInputMode = UsIM
		c.IBflags = IBUsStdFlags
		return &c
	}

	data, err := ioutil.ReadFile(GetConfigPath(engineName))
	if err == nil {
		json.Unmarshal(data, &c)
	}

	return &c
}

func SaveConfig(c *Config, engineName string) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return
	}

	err = ioutil.WriteFile(fmt.Sprintf(configFile, GetConfigDir(engineName), engineName), data, 0644)
	if err != nil {
		log.Println(err)
	}

}
