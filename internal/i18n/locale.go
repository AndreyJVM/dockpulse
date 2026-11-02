package i18n

import (
	"os"
	"strings"
)

// DetectLanguage определяет язык интерфейса на основе переменных окружения.
// Приоритет: DOCKPULSE_LANG -> LC_ALL -> LC_MESSAGES -> LANG.
// Если локаль начинается с "ru", возвращает "ru", иначе по умолчанию "en".
func DetectLanguage() string {
	if custom := strings.TrimSpace(os.Getenv("DOCKPULSE_LANG")); custom != "" {
		if strings.EqualFold(custom, "ru") || strings.HasPrefix(strings.ToLower(custom), "ru") {
			return "ru"
		}
		return "en"
	}

	for _, env := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		val := strings.ToLower(strings.TrimSpace(os.Getenv(env)))
		if strings.HasPrefix(val, "ru_") || strings.HasPrefix(val, "ru.") || val == "ru" {
			return "ru"
		}
	}

	return "en"
}
