package setup

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// GetOpenAIKeyCmd is a Cobra command to fetch and print the OpenAI API key.
var GetOpenAIKeyCmd = &cobra.Command{
	Use:   "get-openai-key",
	Short: "Show the currently stored OpenAI API key",
	Run: func(cmd *cobra.Command, args []string) {
		key := getOpenAIKey()
		if key == "" {
			color.New(color.FgRed, color.Bold).Println("No OpenAI API key found.")
		} else {
			color.New(color.FgGreen, color.Bold).Printf("Current OpenAI API key: %s\n", key)
		}
	},
}

// Step represents a single setup step from the YAML config.
type Step struct {
	Desc string `yaml:"desc"`
	Cmd  string `yaml:"cmd"`
	Meta string `yaml:"meta,omitempty"`
	Mode string `yaml:"mode,omitempty"`
}

// Config represents the overall YAML configuration.
type Config struct {
	Setup []Step `yaml:"setup"`
}

// FixResponse represents the structure of a fix suggestion from OpenAI.
type FixResponse struct {
	Fix         string `json:"fix"`
	Explanation string `json:"explanation"`
}

// RunCmd is the main Cobra command for running the setup. It loads the YAML configuration and executes all setup steps defined in zup.yaml.
var configPathFlag string
var zupServiceName string

func getBaseDirectory() string {
	fullPath, err := os.Getwd()
	if err == nil {
		return filepath.Base(fullPath)
	}
	return ""
}

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run setup steps defined in zup.yaml",
	Run: func(cmd *cobra.Command, args []string) {
		configPath := "zup.yaml"
		if configPathFlag != "" {
			configPath = configPathFlag
		}
		if zupServiceName != "" {
			configPath = getGlobalConfigPathForZupService(zupServiceName)
		}
		if configPathFlag == "auto" {
			configPath = getGlobalConfigPathForZupService(getBaseDirectory())
		}
		runSetup(configPath)
	},
}

func init() {
	RunCmd.Flags().StringVar(&configPathFlag, "path", "", "Path to the config file (default: zup.yaml)")
	RunCmd.Flags().StringVarP(&zupServiceName, "service", "s", "", "Specify the service name to run")
}

// SetOpenAIKeyCmd is a Cobra command to set/update the OpenAI API key globally for the user.
var SetOpenAIKeyCmd = &cobra.Command{
	Use:   "set-openai-key",
	Short: "Set or update your OpenAI API key globally",
	Run: func(cmd *cobra.Command, args []string) {
		setAndStoreOpenAIKeyInteractive()
	},
}

// getOpenAIKey loads the OpenAI API key from env or global config (~/.zup/config.yaml). Returns empty string if not found.
func getOpenAIKey() string {
	// 1. Check env
	key := os.Getenv("OPENAI_API_TOKEN")
	if key != "" {
		return key
	}
	// 2. Check global config
	globalConfigPath := getGlobalConfigPath()
	if data, err := os.ReadFile(globalConfigPath); err == nil {
		type config struct {
			OpenAIToken string `yaml:"openai_api_token"`
		}
		var cfg config
		if err := yaml.Unmarshal(data, &cfg); err == nil && cfg.OpenAIToken != "" {
			return cfg.OpenAIToken
		}
	}
	// 3. Check local config (for backward compatibility)
	localConfigPath := ".zup/config.yaml"
	if data, err := os.ReadFile(localConfigPath); err == nil {
		type config struct {
			OpenAIToken string `yaml:"openai_api_token"`
		}
		var cfg config
		if err := yaml.Unmarshal(data, &cfg); err == nil && cfg.OpenAIToken != "" {
			return cfg.OpenAIToken
		}
	}
	return ""
}

// setAndStoreOpenAIKeyInteractive prompts the user for the OpenAI API key and stores it in ~/.zup/config.yaml
func setAndStoreOpenAIKeyInteractive() {
	color.New(color.FgHiMagenta, color.Bold).Print("Enter your OpenAI API key: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	key := strings.TrimSpace(scanner.Text())
	if key == "" {
		color.New(color.FgRed, color.Bold).Println("No key entered. Aborting.")
		return
	}
	if err := storeOpenAIKey(key); err != nil {
		color.New(color.FgRed, color.Bold).Printf("Failed to store key: %v\n", err)
		return
	}
	color.New(color.FgGreen, color.Bold).Println("OpenAI API key saved globally!")
}

// storeOpenAIKey writes the OpenAI API key to ~/.zup/config.yaml
func storeOpenAIKey(key string) error {
	// Always store in global config
	configDir := getGlobalConfigDir()
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}
	configPath := getGlobalConfigPath()
	type config struct {
		OpenAIToken string `yaml:"openai_api_token"`
	}
	cfg := config{OpenAIToken: key}
	out, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, out, 0600)
}

func getGlobalConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".zup" // fallback
	}
	return home + "/.zup"
}

func getGlobalConfigPath() string {
	return getGlobalConfigDir() + "/config.yaml"
}

func getGlobalConfigPathForZupService(serviceName string) string {
	return getGlobalConfigPath() + "/" + serviceName + ".yaml"
}

/*
runSetup is the entry point for executing the setup process as defined in the YAML configuration file (zup.yaml).

This function attempts to load the configuration file, parse its contents into a Config struct, and then iterates over each setup step defined in the file. For each step, it delegates execution to the executeStep function, which handles command execution and error recovery. If the configuration file cannot be loaded or parsed, an error message is printed and the setup process is aborted.
*/
func runSetup(configPath string) {
	cfg, err := loadConfig(configPath)
	if err != nil {
		fmt.Println("Failed to load config file:", err)
		return
	}
	for _, step := range cfg.Setup {
		executeStep(step)
	}
}

/*
loadConfig reads the YAML configuration from the specified file path and unmarshals it into a Config struct.
It returns a pointer to the Config struct and a possible error. If the file cannot be read (e.g., due to missing file or permissions) or if the YAML is invalid, an error is returned. This function is responsible for ensuring that the setup steps are loaded into memory before execution begins.
*/
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

/*
executeStep is responsible for running a single setup step as defined in the configuration.
It prints the step's description and command to the terminal for user visibility. The function then attempts to execute the command using fixAndRunCommandWithMeta, which handles both normal execution and error recovery. If the command fails and cannot be fixed, an error message is displayed. This function ensures that each step is clearly communicated to the user and that failures are handled gracefully.
*/
func executeStep(step Step) {
	mode := step.Mode
	if mode == "" {
		mode = "same-terminal"
	}
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	fmt.Printf("\n%s %s\n%s %s\n",
		cyan("🔧 Step:"), step.Desc,
		cyan("Command:"), step.Cmd,
	)
	if err := fixAndRunCommandWithMeta(step.Cmd, step.Meta, mode); err != nil {
		color.New(color.FgRed, color.Bold).Printf("\n❌ Command ultimately failed after all fixes: %v\n", err)
	}
}

/*
fixAndRunCommandWithMeta attempts to execute a shell command in the specified mode (e.g., same-terminal or background).
If the command fails, it queries OpenAI for a suggested fix, presents the fix and its explanation to the user, and prompts the user to apply the fix. If the user agrees, the fix is applied recursively until the command succeeds or the user declines further fixes. This function is central to the tool's self-healing capability, allowing for interactive troubleshooting and automated recovery from common errors.
*/
func fixAndRunCommandWithMeta(command, meta, mode string) error {
	err := runCommandWithMode(command, mode)
	if err == nil {
		return nil
	}
	color.New(color.FgRed, color.Bold).Printf("\n❌ Command failed: %s\n", err.Error())
	errMsg := err.Error()
	// Ensure OpenAI key is present before calling getFixFromOpenAIWithMeta
	openaiKey := getOpenAIKey()
	if openaiKey == "" {
		color.New(color.FgHiRed, color.Bold).Println("\n🔑 OpenAI API key not found.")
		setAndStoreOpenAIKeyInteractive()
		openaiKey = getOpenAIKey()
		if openaiKey == "" {
			color.New(color.FgRed, color.Bold).Println("No OpenAI API key set. Aborting fix suggestion.")
			return err
		}
	}
	fix, explanation := getFixFromOpenAIWithMeta(command, errMsg, meta)
	color.New(color.FgYellow, color.Bold).Printf("\n💡 Suggested Fix: %s\n", fix)
	color.New(color.FgHiBlack).Printf("📝 %s\n", explanation)
	if askYesNo(color.New(color.FgGreen, color.Bold).Sprintf("Apply this fix?")) {
		if fixErr := fixAndRunCommandWithMeta(fix, meta, ""); fixErr == nil {
			if mode == "background" {
				if !waitForBinary(getBinaryName(command), 10, time.Second) {
					color.New(color.FgRed, color.Bold).Printf("\n❌ Binary '%s' still not found after fix. Please ensure it is installed and in your PATH.\n", getBinaryName(command))
					return fmt.Errorf("binary '%s' still not found after fix", getBinaryName(command))
				}
			}
			color.New(color.FgGreen, color.Bold).Println("\n✅ Fix applied. Retrying original command...")
			return fixAndRunCommandWithMeta(command, meta, mode)
		} else {
			color.New(color.FgRed, color.Bold).Printf("\n❌ Fix command failed: %v\n", fixErr)
			return fixErr
		}
	}
	return err
}

/*
runCommandWithMode executes a shell command according to the specified mode.
If the mode is 'background', the command is run using nohup so it continues running after the terminal closes, and output is redirected to a log file. The function checks for the existence of the required binary before attempting execution. In the default mode, the command is run in the current terminal session. Errors are returned if the binary is missing or the command fails. This function abstracts the details of command execution modes for the rest of the setup process.
*/
func runCommandWithMode(command, mode string) error {
	switch mode {
	case "background":
		binary := getBinaryName(command)
		if binary == "" {
			return fmt.Errorf("could not determine binary for background command: %s", command)
		}
		if _, err := exec.LookPath(binary); err != nil {
			return fmt.Errorf("binary '%s' not found: %w", binary, err)
		}
		color.New(color.FgCyan).Printf("\n🚀 Running '%s' in background...\n", command)
		// Write the command to a temporary shell script
		tmpFile, err := os.CreateTemp("", "zup-bg-*.sh")
		if err != nil {
			return fmt.Errorf("failed to create temp script: %w", err)
		}
		defer tmpFile.Close()
		// Use login shell to ensure PATH and environment are loaded
		scriptContent := fmt.Sprintf("#!/bin/zsh -l\n%s\n", command)
		if _, err := tmpFile.WriteString(scriptContent); err != nil {
			return fmt.Errorf("failed to write to temp script: %w", err)
		}
		if err := tmpFile.Chmod(0755); err != nil {
			return fmt.Errorf("failed to chmod temp script: %w", err)
		}
		cmd := exec.Command("open", "-a", "Terminal", tmpFile.Name())
		return cmd.Run()
	default:
		return runCommand(command, false)
	}
}

/*
runCommand executes a shell command using bash, with optional output suppression.
If suppressOutput is true, both stdout and stderr are captured and not printed to the terminal; otherwise, output is streamed directly to the terminal. The function returns an error if the command fails, including any captured output for debugging. This function provides a flexible way to run shell commands and handle their output as needed by the setup process.
*/
func runCommand(command string, suppressOutput bool) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdin = os.Stdin
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if !suppressOutput {
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		if suppressOutput {
			return errors.New(strings.TrimSpace(stdout.String() + "\n" + stderr.String()))
		}
		return errors.New(stderr.String())
	}
	if !suppressOutput && stdout.Len() > 0 {
		fmt.Print(stdout.String())
	}
	return nil
}

/*
askYesNo prompts the user with a yes/no question and waits for input from stdin.
The function returns true if the user responds with 'y' or 'yes' (case-insensitive), and false for any other response. This is used to confirm user intent before applying potentially impactful fixes or changes during the setup process.
*/
func askYesNo(prompt string) bool {
	color.New(color.FgHiMagenta, color.Bold).Printf("%s (y/n): ", prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	resp := strings.ToLower(scanner.Text())
	return resp == "y" || resp == "yes"
}

/*
getBinaryName extracts the first word from a shell command string, which is assumed to be the binary or executable name.
If the command string is empty, it returns an empty string. This utility is used to check for the presence of required binaries before attempting to run commands, especially in background mode.
*/
func getBinaryName(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

/*
waitForBinary repeatedly checks if a given binary is available in the system PATH.
It retries up to maxTries times, waiting for the specified delay between attempts. Returns true if the binary is found within the allotted attempts, or false otherwise. This is useful for waiting on installations or updates to complete before proceeding with dependent steps.
*/
func waitForBinary(binary string, maxTries int, delay time.Duration) bool {
	for i := 0; i < maxTries; i++ {
		if _, err := exec.LookPath(binary); err == nil {
			return true
		}
		time.Sleep(delay)
	}
	return false
}
