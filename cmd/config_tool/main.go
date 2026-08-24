package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"ibus-bamboo/config"
)

const EngineName = "bamboo"

func printUsage() {
	fmt.Println("Bamboo Viet Configuration Utility (bamboo-viet-config)")
	fmt.Println("Usage:")
	fmt.Println("  bamboo-viet-config status                   - Display current configuration")
	fmt.Println("  bamboo-viet-config set-method <Telex|VNI>   - Set default input method")
	fmt.Println("  bamboo-viet-config set-mode <1-6|mode_name> - Set default input mode")
	fmt.Println("  bamboo-viet-config set-app <wm_class> <mode>- Set per-app input mode")
	fmt.Println("  bamboo-viet-config list-presets             - List built-in optimal app presets")
	fmt.Println("  bamboo-viet-config restart                  - Restart ibus daemon")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := strings.ToLower(os.Args[1])
	cfg := config.LoadConfig(EngineName)

	switch command {
	case "status":
		fmt.Printf("=== Bamboo Viet Configuration Status ===\n")
		fmt.Printf("Input Method       : %s\n", cfg.InputMethod)
		fmt.Printf("Output Charset     : %s\n", cfg.OutputCharset)
		fmt.Printf("Default Input Mode : [%d] %s\n", cfg.DefaultInputMode, config.ImLookupTable[cfg.DefaultInputMode])
		fmt.Printf("Config File Path   : %s\n", config.GetConfigPath(EngineName))
		fmt.Printf("Macro File Path    : %s\n", config.GetMacroPath(EngineName))

		fmt.Printf("\nPer-App Mappings (%d configured):\n", len(cfg.InputModeMapping))
		if len(cfg.InputModeMapping) == 0 {
			fmt.Println("  (No manual overrides, using built-in presets)")
		} else {
			for app, mode := range cfg.InputModeMapping {
				fmt.Printf("  - %-30s -> [%d] %s\n", app, mode, config.ImLookupTable[mode])
			}
		}

	case "set-method":
		if len(os.Args) < 3 {
			fmt.Println("Error: Missing input method name. Choose: Telex, VNI")
			return
		}
		method := os.Args[2]
		if strings.EqualFold(method, "telex") {
			cfg.InputMethod = "Telex"
		} else if strings.EqualFold(method, "vni") {
			cfg.InputMethod = "VNI"
		} else {
			fmt.Printf("Error: Unsupported method %q. Choose: Telex, VNI\n", method)
			return
		}
		config.SaveConfig(cfg, EngineName)
		fmt.Printf("✓ Input method successfully updated to: %s\n", cfg.InputMethod)

	case "set-mode":
		if len(os.Args) < 3 {
			fmt.Println("Error: Missing mode. Available modes:")
			for id, name := range config.ImLookupTable {
				fmt.Printf("  %d: %s\n", id, name)
			}
			return
		}
		arg := os.Args[2]
		modeID, err := strconv.Atoi(arg)
		if err != nil {
			// Try matching mode name
			found := false
			for id, name := range config.ImLookupTable {
				if strings.Contains(strings.ToLower(name), strings.ToLower(arg)) {
					modeID = id
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("Error: Invalid mode %q\n", arg)
				return
			}
		}

		if config.ImLookupTable[modeID] == "" {
			fmt.Printf("Error: Mode ID %d does not exist\n", modeID)
			return
		}

		cfg.DefaultInputMode = modeID
		config.SaveConfig(cfg, EngineName)
		fmt.Printf("✓ Default input mode updated to: [%d] %s\n", modeID, config.ImLookupTable[modeID])

	case "set-app":
		if len(os.Args) < 4 {
			fmt.Println("Usage: bamboo-viet-config set-app <wm_class> <mode_id>")
			return
		}
		app := os.Args[2]
		modeID, err := strconv.Atoi(os.Args[3])
		if err != nil || config.ImLookupTable[modeID] == "" {
			fmt.Println("Error: Invalid mode ID. Choose from 1 to 7.")
			return
		}

		if cfg.InputModeMapping == nil {
			cfg.InputModeMapping = make(map[string]int)
		}
		cfg.InputModeMapping[app] = modeID
		config.SaveConfig(cfg, EngineName)
		fmt.Printf("✓ Set app override: %q -> [%d] %s\n", app, modeID, config.ImLookupTable[modeID])

	case "list-presets":
		fmt.Println("=== Built-in Optimal App Presets ===")
		for app, mode := range config.DefaultAppPresets {
			fmt.Printf("  %-30s -> [%d] %s\n", app, mode, config.ImLookupTable[mode])
		}

	case "restart":
		fmt.Println("Restarting ibus-daemon...")
		cmd := exec.Command("ibus", "restart")
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Note: 'ibus restart' returned %v (you can run 'ibus-daemon -drx' manually if needed)\n", err)
		} else {
			fmt.Println("✓ IBus restarted successfully.")
		}

	default:
		printUsage()
	}
}
