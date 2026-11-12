package i18n

import "testing"

func TestGetReturnsEnglishDictionary(t *testing.T) {
	dict := Get()

	if dict.TabHost != "1: Host Overview" {
		t.Fatalf("unexpected host tab label: %q", dict.TabHost)
	}
	if dict.StatusRunning != "Running" {
		t.Fatalf("unexpected running status: %q", dict.StatusRunning)
	}
	if dict.HelpTitle != "HOTKEYS:" {
		t.Fatalf("unexpected help title: %q", dict.HelpTitle)
	}
}
