package components

import (
	"math"
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
		got := Truncate(c.input, c.maxLen)
		if got != c.expected {
			t.Errorf("Truncate(%q, %d) = %q; want %q", c.input, c.maxLen, got, c.expected)
		}
	}
}

func TestClamp(t *testing.T) {
	if val := Clamp(-5.0, 0.0, 100.0); val != 0.0 {
		t.Errorf("expected 0.0, got %f", val)
	}
	if val := Clamp(120.0, 0.0, 100.0); val != 100.0 {
		t.Errorf("expected 100.0, got %f", val)
	}
	if val := Clamp(42.5, 0.0, 100.0); val != 42.5 {
		t.Errorf("expected 42.5, got %f", val)
	}
	if val := Clamp(math.NaN(), 0.0, 100.0); val != 0.0 {
		t.Errorf("expected 0.0 for NaN, got %f", val)
	}
}

func TestPercentStyle(t *testing.T) {
	cases := []struct {
		val       float64
		wantColor string
	}{
		{20.0, "green"},
		{59.9, "green"},
		{60.0, "yellow"},
		{79.9, "yellow"},
		{80.0, "red"},
		{100.0, "red"},
	}

	for _, c := range cases {
		style := PercentStyle(c.val)
		rendered := style.Render("test")
		if rendered == "test" && style.GetForeground() == nil {
			t.Errorf("expected colored output for %f, got unstyled", c.val)
		}
	}
}

func TestFormatPercent(t *testing.T) {
	p1 := FormatPercent(25.0)
	if !strings.Contains(p1, "25.0%") {
		t.Errorf("expected '25.0%%' in %q", p1)
	}

	p2 := FormatPercent(65.0)
	if !strings.Contains(p2, "65.0%") {
		t.Errorf("expected '65.0%%' in %q", p2)
	}

	p3 := FormatPercent(92.0)
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
		{50.0, 0},  // edge case: zero width
		{50.0, -2}, // edge case: negative width
		{50.0, 1},  // minimal width
	}

	for _, c := range cases {
		bar := FormatBar(c.pct, c.width)
		if !strings.HasPrefix(bar, "[") || !strings.Contains(bar, "]") || !strings.Contains(bar, "%") {
			t.Errorf("FormatBar(%f, %d) returned invalid format: %q", c.pct, c.width, bar)
		}
	}
}

func TestRenderBar(t *testing.T) {
	// Without percent label
	noPct := RenderBar(50.0, 10, false)
	if strings.Contains(noPct, "%") {
		t.Errorf("expected no percent in %q", noPct)
	}
	if !strings.HasPrefix(noPct, "[") || !strings.HasSuffix(noPct, "]") {
		t.Errorf("expected bracketed bar, got %q", noPct)
	}

	// Zero width without percent
	zeroNoPct := RenderBar(50.0, 0, false)
	if zeroNoPct != "[]" {
		t.Errorf("expected '[]' for zero width without percent, got %q", zeroNoPct)
	}
}

func TestRenderSparkline(t *testing.T) {
	t.Run("empty history", func(t *testing.T) {
		if res := RenderSparkline(nil); res != "" {
			t.Errorf("expected empty string for nil history, got %q", res)
		}
		if res := RenderSparkline([]float64{}); res != "" {
			t.Errorf("expected empty string for empty history, got %q", res)
		}
	})

	t.Run("boundary and clamped values", func(t *testing.T) {
		history := []float64{-10.0, 0.0, 50.0, 100.0, 150.0}
		res := RenderSparkline(history)
		if res == "" {
			t.Fatal("expected non-empty rendered sparkline")
		}
	})

	t.Run("color grading checks", func(t *testing.T) {
		low := RenderSparkline([]float64{10.0})
		if !strings.Contains(low, string(SparkChars[0])) {
			t.Errorf("expected low sparkline character, got %q", low)
		}

		med := RenderSparkline([]float64{70.0})
		if !strings.Contains(med, string(SparkChars[5])) {
			t.Errorf("expected medium sparkline character, got %q", med)
		}

		high := RenderSparkline([]float64{95.0})
		if !strings.Contains(high, string(SparkChars[7])) {
			t.Errorf("expected high sparkline character, got %q", high)
		}
	})
}
