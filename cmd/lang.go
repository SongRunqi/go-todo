package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/SongRunqi/go-todo/internal/i18n"
)

// Language represents a supported language
type Language struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	NativeName string `json:"native_name"`
}

var supportedLanguages = []Language{
	{Code: "en", Name: "English", NativeName: "English"},
	{Code: "zh", Name: "Chinese", NativeName: "中文"},
}

// langCmd represents the lang command
var langCmd = &cobra.Command{
	Use:   "lang",
	Short: "",
	Long:  "",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Override parent's PersistentPreRun - lang command doesn't need todos
		// We still need to initialize i18n though (already done in init)
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Default: list languages
		listLanguages()
	},
}

// langListCmd represents the "lang list" command
var langListCmd = &cobra.Command{
	Use:   "list",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		listLanguages()
	},
}

// langSetCmd represents the "lang set" command
var langSetCmd = &cobra.Command{
	Use:   "set <language-code>",
	Short: "",
	Long:  "",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		langCode := args[0]

		// Validate language code
		validLang := false
		for _, lang := range supportedLanguages {
			if lang.Code == langCode {
				validLang = true
				break
			}
		}

		if !validLang {
			fmt.Fprintf(os.Stderr, i18n.T("cmd.lang.error.unsupported_code"), langCode)
			fmt.Fprintf(os.Stderr, i18n.T("cmd.lang.error.supported_languages"))
			os.Exit(1)
		}

		// Save language preference to config file
		if err := saveLanguageConfig(langCode); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cmd.lang.error.save_config"), err)
			os.Exit(1)
		}

		// Set language for current session
		if err := i18n.SetLanguage(langCode); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cmd.lang.error.set_language"), err)
			os.Exit(1)
		}

		// Output success message in the selected language
		successMsg := map[string]string{
			"en": i18n.T("cmd.lang.success.language_set_en"),
			"zh": i18n.T("cmd.lang.success.language_set_zh"),
		}
		fmt.Println(successMsg[langCode])
	},
}

// langCurrentCmd represents the "lang current" command
var langCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		currentLang := i18n.GetLanguage()
		for _, lang := range supportedLanguages {
			if lang.Code == currentLang {
				fmt.Printf(i18n.T("cmd.lang.current_language"), lang.NativeName, lang.Code)
				return
			}
		}
		fmt.Printf(i18n.T("cmd.lang.current_language_code"), currentLang)
	},
}

func listLanguages() {
	currentLang := i18n.GetLanguage()

	// Create Alfred-compatible JSON output
	type AlfredItem struct {
		Title        string `json:"title"`
		Subtitle     string `json:"subtitle"`
		Arg          string `json:"arg"`
		Autocomplete string `json:"autocomplete"`
		Icon         string `json:"icon,omitempty"`
	}

	type AlfredOutput struct {
		Items []AlfredItem `json:"items"`
	}

	var items []AlfredItem

	for _, lang := range supportedLanguages {
		title := fmt.Sprintf("%s (%s)", lang.NativeName, lang.Name)
		subtitle := fmt.Sprintf("Language code: %s", lang.Code)

		// Add indicator for current language
		if lang.Code == currentLang {
			title = "✓ " + title + " [Current]"
			subtitle = "Currently selected - " + subtitle
		}

		items = append(items, AlfredItem{
			Title:        title,
			Subtitle:     subtitle,
			Arg:          lang.Code,
			Autocomplete: lang.Code,
		})
	}

	output := AlfredOutput{Items: items}
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonData))
}

func saveLanguageConfig(langCode string) error {
	// Get config directory (use the same as todo files)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	configDir := filepath.Join(homeDir, ".todo")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config.json")

	// Read existing config or create new one
	type Config struct {
		Language string `json:"language"`
	}

	cfg := Config{Language: langCode}

	// Write config file
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(langCmd)
	langCmd.AddCommand(langListCmd)
	langCmd.AddCommand(langSetCmd)
	langCmd.AddCommand(langCurrentCmd)
}
