package tui

import (
	"strings"
	"testing"
)

func TestTruncate(t *testing.T) {
	cases := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 8, "hello..."},
		{"привет мир", 9, "привет..."},
		{"абвгдеж", 3, "абв"},
		{"абв", 2, "аб"},
		{"", 5, ""},
		{"short", 0, ""},
	}

	for _, c := range cases {
		got := truncate(c.input, c.maxLen)
		if got != c.expected {
			t.Errorf("truncate(%q, %d) = %q; want %q", c.input, c.maxLen, got, c.expected)
		}
	}
}

func TestFormatPercent(t *testing.T) {
	p1 := formatPercent(25.0)
	if !strings.Contains(p1, "25.0%") {
		t.Errorf("expected '25.0%%' in %q", p1)
	}

	p2 := formatPercent(65.0)
	if !strings.Contains(p2, "65.0%") {
		t.Errorf("expected '65.0%%' in %q", p2)
	}

	p3 := formatPercent(92.0)
	if !strings.Contains(p3, "92.0%") {
		t.Errorf("expected '92.0%%' in %q", p3)
	}
}

func TestFormatBar(t *testing.T) {
	cases := []struct {
		pct   float64
		width int
	}{
		{-5.0, 10},
		{0.0, 10},
		{50.0, 20},
		{75.0, 15},
		{90.0, 15},
		{110.0, 10},
	}

	for _, c := range cases {
		bar := formatBar(c.pct, c.width)
		if !strings.HasPrefix(bar, "[") || !strings.Contains(bar, "]") || !strings.Contains(bar, "%") {
			t.Errorf("formatBar(%f, %d) returned invalid format: %q", c.pct, c.width, bar)
		}
	}
}

func TestRenderSparkline(t *testing.T) {
	t.Run("empty history", func(t *testing.T) {
		if res := renderSparkline(nil); res != "" {
			t.Errorf("expected empty string for nil history, got %q", res)
		}
		if res := renderSparkline([]float64{}); res != "" {
			t.Errorf("expected empty string for empty history, got %q", res)
		}
	})

	t.Run("boundary and clamped values", func(t *testing.T) {
		history := []float64{-10.0, 0.0, 50.0, 100.0, 150.0}
		res := renderSparkline(history)
		if res == "" {
			t.Fatal("expected non-empty rendered sparkline")
		}
	})

	t.Run("color grading checks", func(t *testing.T) {
		low := renderSparkline([]float64{10.0})
		if !strings.Contains(low, string(sparkChars[0])) {
			t.Errorf("expected low sparkline character, got %q", low)
		}

		med := renderSparkline([]float64{70.0})
		if !strings.Contains(med, string(sparkChars[5])) {
			t.Errorf("expected medium sparkline character, got %q", med)
		}

		high := renderSparkline([]float64{95.0})
		if !strings.Contains(high, string(sparkChars[7])) {
			t.Errorf("expected high sparkline character, got %q", high)
		}
	})
}
