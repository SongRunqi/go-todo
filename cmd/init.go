package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/SongRunqi/go-todo/internal/i18n"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "",
	Long:  "",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Override parent's PersistentPreRun - init command doesn't need todos
	},
	Run: func(cmd *cobra.Command, args []string) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get home directory: %v\n", err)
			os.Exit(1)
		}

		todoDir := filepath.Join(homeDir, ".todo")

		// Check if directory already exists
		if _, err := os.Stat(todoDir); err == nil {
			fmt.Printf("✓ Todo directory already exists: %s\n", todoDir)
		} else {
			// Create .todo directory
			if err := os.MkdirAll(todoDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to create todo directory: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Created todo directory: %s\n", todoDir)
		}

		// Initialize todo.json if it doesn't exist
		todoFile := filepath.Join(todoDir, "todo.json")
		if _, err := os.Stat(todoFile); os.IsNotExist(err) {
			emptyTodos := []interface{}{}
			data, _ := json.MarshalIndent(emptyTodos, "", "  ")
			if err := os.WriteFile(todoFile, data, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to create todo.json: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Created todo file: %s\n", todoFile)
		} else {
			fmt.Printf("✓ Todo file already exists: %s\n", todoFile)
		}

		// Initialize todo_back.json if it doesn't exist
		backupFile := filepath.Join(todoDir, "todo_back.json")
		if _, err := os.Stat(backupFile); os.IsNotExist(err) {
			emptyTodos := []interface{}{}
			data, _ := json.MarshalIndent(emptyTodos, "", "  ")
			if err := os.WriteFile(backupFile, data, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to create todo_back.json: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Created backup file: %s\n", backupFile)
		} else {
			fmt.Printf("✓ Backup file already exists: %s\n", backupFile)
		}

		// Initialize config.json with language selection
		configFile := filepath.Join(todoDir, "config.json")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			fmt.Println("\nLanguage selection / 语言选择:")
			fmt.Println("  1. English")
			fmt.Println("  2. 中文")
			fmt.Print("\nSelect language (1-2) [1]: ")

			var choice string
			fmt.Scanln(&choice)

			lang := "en"
			if choice == "2" {
				lang = "zh"
			}

			config := map[string]string{
				"language": lang,
			}
			data, _ := json.MarshalIndent(config, "", "  ")
			if err := os.WriteFile(configFile, data, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to create config.json: %v\n", err)
				os.Exit(1)
			}

			// Set the language for current session
			i18n.SetLanguage(lang)

			fmt.Printf("\n✓ Created config file: %s\n", configFile)
			if lang == "zh" {
				fmt.Println("✓ 语言已设置为中文")
			} else {
				fmt.Println("✓ Language set to English")
			}
		} else {
			fmt.Printf("✓ Config file already exists: %s\n", configFile)
		}

		fmt.Println("\n╔════════════════════════════════════════╗")
		fmt.Println("║  Initialization Complete! 🎉           ║")
		fmt.Println("╚════════════════════════════════════════╝")
		fmt.Println("\nYou can now use todo-go:")
		fmt.Println("  • List tasks:        todo list")
		fmt.Println("  • Create task:       todo \"买菜 明天截止\"")
		fmt.Println("  • Get help:          todo --help")
		fmt.Println("  • Change language:   todo lang set zh/en")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
