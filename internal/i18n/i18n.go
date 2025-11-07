package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

//go:embed translations/*.json
var translationsFS embed.FS

var (
	currentLang = "en"
	translations = make(map[string]map[string]string)
	mu sync.RWMutex
)

// Init initializes the i18n system with the specified language
func Init(lang string) error {
	mu.Lock()
	defer mu.Unlock()

	if lang == "" {
		lang = detectLanguage()
	}

	// Normalize language code
	lang = strings.ToLower(lang)
	if strings.HasPrefix(lang, "zh") {
		lang = "zh"
	} else if strings.HasPrefix(lang, "en") {
		lang = "en"
	}

	currentLang = lang

	// Load English as fallback
	if err := loadLanguage("en"); err != nil {
		return fmt.Errorf("failed to load English translations: %w", err)
	}

	// Load the requested language if it's not English
	if lang != "en" {
		if err := loadLanguage(lang); err != nil {
			return fmt.Errorf("failed to load %s translations: %w", lang, err)
		}
	}

	return nil
}

// loadLanguage loads translations for a specific language
func loadLanguage(lang string) error {
	filename := fmt.Sprintf("translations/%s.json", lang)
	data, err := translationsFS.ReadFile(filename)
	if err != nil {
		return err
	}

	var langTranslations map[string]string
	if err := json.Unmarshal(data, &langTranslations); err != nil {
		return err
	}

	translations[lang] = langTranslations
	return nil
}

// detectLanguage detects the user's language from environment variables
func detectLanguage() string {
	// Check LANGUAGE, LC_ALL, LC_MESSAGES, LANG in order
	for _, env := range []string{"LANGUAGE", "LC_ALL", "LC_MESSAGES", "LANG"} {
		if lang := os.Getenv(env); lang != "" {
			return lang
		}
	}
	return "en"
}

// T translates a key to the current language
func T(key string, args ...interface{}) string {
	mu.RLock()
	defer mu.RUnlock()

	// Try current language first
	if langMap, ok := translations[currentLang]; ok {
		if msg, ok := langMap[key]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(msg, args...)
			}
			return msg
		}
	}

	// Fallback to English
	if currentLang != "en" {
		if langMap, ok := translations["en"]; ok {
			if msg, ok := langMap[key]; ok {
				if len(args) > 0 {
					return fmt.Sprintf(msg, args...)
				}
				return msg
			}
		}
	}

	// If all else fails, return the key itself
	return key
}

// SetLanguage changes the current language
func SetLanguage(lang string) error {
	mu.Lock()
	defer mu.Unlock()

	// Normalize language code
	lang = strings.ToLower(lang)
	if strings.HasPrefix(lang, "zh") {
		lang = "zh"
	} else if strings.HasPrefix(lang, "en") {
		lang = "en"
	}

	// Check if translations are already loaded
	if _, ok := translations[lang]; !ok {
		if err := loadLanguage(lang); err != nil {
			return err
		}
	}

	currentLang = lang
	return nil
}

// GetLanguage returns the current language
func GetLanguage() string {
	mu.RLock()
	defer mu.RUnlock()
	return currentLang
}

// GetAvailableLanguages returns a list of available languages
func GetAvailableLanguages() []string {
	return []string{"en", "zh"}
}
