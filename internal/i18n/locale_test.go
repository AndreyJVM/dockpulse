package i18n

import (
	"os"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	// Сохраняем исходное окружение
	origDockpulse := os.Getenv("DOCKPULSE_LANG")
	origLCAll := os.Getenv("LC_ALL")
	origLCMsg := os.Getenv("LC_MESSAGES")
	origLang := os.Getenv("LANG")
	defer func() {
		os.Setenv("DOCKPULSE_LANG", origDockpulse)
		os.Setenv("LC_ALL", origLCAll)
		os.Setenv("LC_MESSAGES", origLCMsg)
		os.Setenv("LANG", origLang)
	}()

	// 1. DOCKPULSE_LANG имеет наивысший приоритет
	os.Setenv("DOCKPULSE_LANG", "ru")
	os.Setenv("LANG", "en_US.UTF-8")
	if got := DetectLanguage(); got != "ru" {
		t.Errorf("expected 'ru' from DOCKPULSE_LANG, got %q", got)
	}

	os.Setenv("DOCKPULSE_LANG", "en")
	os.Setenv("LANG", "ru_RU.UTF-8")
	if got := DetectLanguage(); got != "en" {
		t.Errorf("expected 'en' from DOCKPULSE_LANG override, got %q", got)
	}

	// Сбрасываем DOCKPULSE_LANG для проверки стандартных переменных
	os.Unsetenv("DOCKPULSE_LANG")

	// 2. Русская локаль по LANG
	os.Setenv("LANG", "ru_RU.UTF-8")
	os.Unsetenv("LC_ALL")
	os.Unsetenv("LC_MESSAGES")
	if got := DetectLanguage(); got != "ru" {
		t.Errorf("expected 'ru' from LANG=ru_RU.UTF-8, got %q", got)
	}

	// 3. Английская локаль / POSIX / C
	cases := []string{"C", "POSIX", "en_US.UTF-8", "en_GB.UTF-8", "fr_FR.UTF-8", ""}
	for _, c := range cases {
		os.Setenv("LANG", c)
		if got := DetectLanguage(); got != "en" {
			t.Errorf("expected 'en' for LANG=%q, got %q", c, got)
		}
	}
}

func TestGetDictionary(t *testing.T) {
	ru := Get("ru")
	if ru.TabHost != "1: Host Overview" && ru.FooterTabs != "Вкладки" {
		t.Errorf("unexpected ru dictionary content")
	}

	en := Get("en")
	if en.FooterTabs != "Tabs" {
		t.Errorf("unexpected en dictionary content: %q", en.FooterTabs)
	}

	// Любой неизвестный язык должен отдавать английский словарь по умолчанию
	unknown := Get("de")
	if unknown.FooterTabs != "Tabs" {
		t.Errorf("expected fallback to en for unknown lang, got %q", unknown.FooterTabs)
	}
}
