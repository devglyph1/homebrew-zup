package setup

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// --- MODIFICATION START ---
// Updated Config struct to include a new 'Env' field for environment variables.
type Config struct {
	Env   map[string]string `yaml:"env,omitempty"`
	Setup []Step            `yaml:"setup"`
}

// --- MODIFICATION END ---

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

// --- Global Configuration Management ---

/*
loadGlobalConfig reads the global ~/.zup/config.yaml file into a map.
This function serves as a centralized way to access all stored user settings, including API keys and custom variables. It gracefully handles cases where the config file doesn't exist by returning an empty map, allowing the application to proceed without errors.
*/
func loadGlobalConfig() (map[string]interface{}, error) {
	configPath := getGlobalConfigPath()
	config := make(map[string]interface{})

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil // Return empty map if file doesn't exist
		}
		return nil, err // Return other errors
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return config, nil
}

/*
saveGlobalConfig saves a map to the global ~/.zup/config.yaml file.
It ensures the configuration directory exists and then marshals the provided map into YAML format, overwriting the existing configuration file. This allows for persistent storage of any key-value setting.
*/
func saveGlobalConfig(config map[string]interface{}) error {
	configDir := getGlobalConfigDir()
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := getGlobalConfigPath()
	return os.WriteFile(configPath, data, 0600)
}

/*
storeGlobalConfigValue adds or updates a key-value pair in the global config.
It reads the current configuration, updates the specified key with the new value, and saves the configuration back to disk. This is the primary method for persisting user-defined variables and settings.
*/
func storeGlobalConfigValue(key string, value interface{}) error {
	config, err := loadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load global config: %w", err)
	}
	config[key] = value
	return saveGlobalConfig(config)
}

/*
getGlobalConfigValue retrieves a value for a given key from the global config.
It returns an empty string if the key is not found or if the configuration file cannot be read, providing a safe way to access stored values.
*/
func getGlobalConfigValue(key string) string {
	config, err := loadGlobalConfig()
	if err != nil {
		return ""
	}
	if val, ok := config[key].(string); ok {
		return val
	}
	return ""
}

// getOpenAIKey loads the OpenAI API key from env or global config (~/.zup/config.yaml). Returns empty string if not found.
func getOpenAIKey() string {
	// 1. Check env
	key := os.Getenv("OPENAI_API_TOKEN")
	if key != "" {
		return key
	}
	// 2. Check global config
	key = getGlobalConfigValue("openai_api_token")
	if key != "" {
		return key
	}
	// 3. Check local config (for backward compatibility)
	localConfigPath := ".zup/config.yaml"
	if data, err := os.ReadFile(localConfigPath); err == nil {
		var cfg struct {
			OpenAIToken string `yaml:"openai_api_token"`
		}
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
	return storeGlobalConfigValue("openai_api_token", key)
}

func getGlobalConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".zup" // fallback
	}
	return filepath.Join(home, ".zup")
}

func getGlobalConfigPath() string {
	return filepath.Join(getGlobalConfigDir(), "config.yaml")
}

func getGlobalConfigPathForZupService(serviceName string) string {
	return filepath.Join(getGlobalConfigDir(), serviceName+".yaml")
}

/*
resolveCommandVariables finds variables like ${VAR_NAME} in a command string.
For each unique variable found, it checks the global configuration for a stored value. If no value is found, it prompts the user for input. It now accepts empty strings as valid input and saves them to prevent re-prompting. Finally, it returns the command string with all variable placeholders replaced by their corresponding values.
*/
func resolveCommandVariables(command string) (string, error) {
	re := regexp.MustCompile(`\$\{(.*?)\}`)
	matches := re.FindAllStringSubmatch(command, -1)
	resolvedCommand := command
	uniqueVars := make(map[string]bool)
	config, err := loadGlobalConfig()
	if err != nil {
		return "", fmt.Errorf("could not load global config: %w", err)
	}

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		varName := match[1]
		placeholder := match[0]

		if _, seen := uniqueVars[varName]; seen {
			continue
		}
		uniqueVars[varName] = true

		var value string
		if storedValue, exists := config[varName]; exists {
			value = fmt.Sprintf("%v", storedValue)
		} else {
			// Variable not found, prompt the user.
			magentaBold := color.New(color.FgHiMagenta, color.Bold)
			magentaBold.Printf("Please enter a value for '%s': ", varName)

			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			inputValue := strings.TrimSpace(scanner.Text())

			// Store the value (even if empty) to avoid asking again.
			if err := storeGlobalConfigValue(varName, inputValue); err != nil {
				color.New(color.FgRed, color.Bold).Printf("Warning: Failed to store variable '%s' globally: %v\n", varName, err)
			}
			value = inputValue
			// Update the in-memory config for the current run
			config[varName] = inputValue
		}
		resolvedCommand = strings.ReplaceAll(resolvedCommand, placeholder, value)
	}
	return resolvedCommand, nil
}

// --- MODIFICATION START ---

/*
processEnvVariables reads the 'env' key from the zup.yaml config,
creates or updates the zup.env file with these variables, and logs the process.
*/
func processEnvVariables(configPath string) error {
	cfg, err := loadConfig(configPath)
	if err != nil {
		// If the file doesn't exist, it's not an error for this specific function.
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if len(cfg.Env) == 0 {
		// It's normal for a config to not have an 'env' section.
		return nil
	}

	envFile, err := os.Create("zup.env")
	if err != nil {
		return fmt.Errorf("failed to create zup.env file: %w", err)
	}
	defer envFile.Close()

	color.New(color.FgCyan, color.Bold).Println("\n⚙️  Processing environment variables from zup.yaml...")
	writer := bufio.NewWriter(envFile)
	for key, value := range cfg.Env {
		// Resolve any global variables (e.g., ${MY_SECRET}) within the env value itself.
		resolvedValue, err := resolveCommandVariables(value)
		if err != nil {
			return fmt.Errorf("failed to resolve variable in env value for key '%s': %w", key, err)
		}

		envLine := fmt.Sprintf("export %s=\"%s\"\n", key, resolvedValue)
		_, err = writer.WriteString(envLine)
		if err != nil {
			return fmt.Errorf("failed to write to zup.env: %w", err)
		}
		// To avoid printing secrets, we show a generic message.
		color.New(color.FgGreen).Printf("✓ Exporting '%s' variable\n", key)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer for zup.env: %w", err)
	}

	color.New(color.FgGreen, color.Bold).Println("✅ zup.env file created.")
	return nil
}

/*
runSetup is the entry point for executing the setup process as defined in the YAML configuration file (zup.yaml).
It first processes any environment variables into a temporary 'zup.env' file. It then loads the configuration and iterates over each setup step. The 'zup.env' file is automatically sourced by each command and deleted upon completion.
*/
func runSetup(configPath string) {
	// Process env variables from zup.yaml and create the zup.env file first.
	if err := processEnvVariables(configPath); err != nil {
		color.New(color.FgRed, color.Bold).Printf("❌ Failed to process environment variables: %v\n", err)
		return
	}
	// Defer the cleanup of zup.env to ensure it runs even if a step fails or the program exits.
	defer os.Remove("zup.env")

	cfg, err := loadConfig(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			color.New(color.FgRed, color.Bold).Printf("❌ Config file not found at '%s'. Aborting.\n", configPath)
		} else {
			color.New(color.FgRed, color.Bold).Printf("❌ Failed to load config file '%s': %v\n", configPath, err)
		}
		return
	}

	for _, step := range cfg.Setup {
		executeStep(step)
	}
}

// --- MODIFICATION END ---

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
It first resolves any variables (e.g., ${VAR_NAME}) in the command string, prompting the user for input if necessary. It then prints the step's description and the resolved command to the terminal. Finally, it executes the command using fixAndRunCommandWithMeta, which handles both normal execution and error recovery.
*/
func executeStep(step Step) {
	mode := step.Mode
	if mode == "" {
		mode = "same-terminal"
	}

	resolvedCmd, err := resolveCommandVariables(step.Cmd)
	if err != nil {
		color.New(color.FgRed, color.Bold).Printf("\n❌ Failed to resolve variables: %v\n", err)
		return
	}

	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	fmt.Printf("\n%s %s\n%s %s\n",
		cyan("🔧 Step:"), step.Desc,
		cyan("Command:"), resolvedCmd,
	)
	if err := fixAndRunCommandWithMeta(resolvedCmd, step.Meta, mode); err != nil {
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
	fix, explanation := getFixFromOpenAIWithMeta(command, errMsg, meta, openaiKey)
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

// --- MODIFICATION START ---
/*
runCommand executes a shell command using bash. It ensures that any environment variables
defined in a temporary 'zup.env' file are sourced and available to the command.
Output can be suppressed, and stderr is captured for detailed error reporting.
*/
func runCommand(command string, suppressOutput bool) error {
	// The state file acts as the "memory" for environment variables between executions.
	// It is created and populated by the `processEnvVariables` function.
	const envFile = "zup.env"

	// Construct the command to be run by bash.
	// 1. `source zup.env`: This loads the environment variables from the file.
	// 2. `%s`: This executes the actual command from the zup.yaml.
	// `2>/dev/null` is used to suppress any "file not found" errors if zup.env doesn't exist.
	wrappedCmd := fmt.Sprintf("source %s 2>/dev/null; %s", envFile, command)

	cmd := exec.Command("bash", "-c", wrappedCmd)

	cmd.Stdin = os.Stdin
	var stdout, stderr bytes.Buffer
	cmd.Stderr = &stderr

	if suppressOutput {
		cmd.Stdout = &stdout
	} else {
		cmd.Stdout = os.Stdout
	}

	if err := cmd.Run(); err != nil {
		// The error from cmd.Run() is often just "exit status X".
		// The actual error message is in stderr, which is more useful.
		errMsg := strings.TrimSpace(stderr.String())
		if !suppressOutput && errMsg != "" {
			// If not suppressing, the user hasn't seen the error yet because we captured it.
			// So, print it now before returning.
			fmt.Fprintln(os.Stderr, errMsg)
		}

		if errMsg != "" {
			return errors.New(errMsg)
		}
		// Fallback to the original error if stderr is empty
		return err
	}
	return nil
}

// --- MODIFICATION END ---

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
