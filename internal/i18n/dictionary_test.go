package i18n

import "testing"

func TestGetReturnsEnglishDictionary(t *testing.T) {
	dict := Get()

	if dict.TabDocker != "1: Containers" {
		t.Fatalf("unexpected Docker tab label: %q", dict.TabDocker)
	}
	if dict.StatusRunning != "Running" {
		t.Fatalf("unexpected running status: %q", dict.StatusRunning)
	}
	if dict.HelpTitle != "HOTKEYS:" {
		t.Fatalf("unexpected help title: %q", dict.HelpTitle)
	}
}
