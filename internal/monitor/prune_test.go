package monitor

import (
	"strings"
	"testing"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{50 * 1024 * 1024, "50.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{2500 * 1024 * 1024, "2.4 GB"},
		{-10, "0 B"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := FormatBytes(tt.bytes)
			if got != tt.expected {
				t.Errorf("FormatBytes(%d) = %q; want %q", tt.bytes, got, tt.expected)
			}
		})
	}
}

func TestParseSystemDFRaw(t *testing.T) {
	var raw systemDFRaw

	// 2 images: 1 active (100MB), 1 dangling (50MB)
	raw.Images = []struct {
		Size       int64 `json:"Size"`
		SharedSize int64 `json:"SharedSize"`
		Containers int64 `json:"Containers"`
	}{
		{Size: 100 * 1024 * 1024, Containers: 1},
		{Size: 50 * 1024 * 1024, Containers: 0},
	}

	// 2 containers: 1 running (10MB), 1 stopped (20MB)
	raw.Containers = []struct {
		SizeRw     int64  `json:"SizeRw"`
		SizeRootFs int64  `json:"SizeRootFs"`
		State      string `json:"State"`
	}{
		{SizeRw: 10 * 1024 * 1024, State: "running"},
		{SizeRw: 20 * 1024 * 1024, State: "exited"},
	}

	// 2 volumes: 1 active (30MB), 1 unused (40MB)
	raw.Volumes = []struct {
		UsageData struct {
			Size     int64 `json:"Size"`
			RefCount int64 `json:"RefCount"`
		} `json:"UsageData"`
	}{
		{UsageData: struct {
			Size     int64 `json:"Size"`
			RefCount int64 `json:"RefCount"`
		}{Size: 30 * 1024 * 1024, RefCount: 1}},
		{UsageData: struct {
			Size     int64 `json:"Size"`
			RefCount int64 `json:"RefCount"`
		}{Size: 40 * 1024 * 1024, RefCount: 0}},
	}

	// 1 build cache item: reclaimable (60MB)
	raw.BuildCache = []struct {
		Size        int64 `json:"Size"`
		Reclaimable bool  `json:"Reclaimable"`
	}{
		{Size: 60 * 1024 * 1024, Reclaimable: true},
	}

	usage := ParseSystemDFRaw(raw)

	// Check Images
	if usage.Images.TotalCount != 2 || usage.Images.ActiveCount != 1 {
		t.Errorf("Images count mismatch: %+v", usage.Images)
	}
	if usage.Images.TotalSize != 150*1024*1024 || usage.Images.Reclaimable != 50*1024*1024 {
		t.Errorf("Images size mismatch: %+v", usage.Images)
	}

	// Check Containers
	if usage.Containers.TotalCount != 2 || usage.Containers.ActiveCount != 1 {
		t.Errorf("Containers count mismatch: %+v", usage.Containers)
	}
	if usage.Containers.TotalSize != 30*1024*1024 || usage.Containers.Reclaimable != 20*1024*1024 {
		t.Errorf("Containers size mismatch: %+v", usage.Containers)
	}

	// Check Volumes
	if usage.Volumes.TotalCount != 2 || usage.Volumes.ActiveCount != 1 {
		t.Errorf("Volumes count mismatch: %+v", usage.Volumes)
	}
	if usage.Volumes.TotalSize != 70*1024*1024 || usage.Volumes.Reclaimable != 40*1024*1024 {
		t.Errorf("Volumes size mismatch: %+v", usage.Volumes)
	}

	// Check BuildCache
	if usage.BuildCache.TotalCount != 1 || usage.BuildCache.Reclaimable != 60*1024*1024 {
		t.Errorf("BuildCache mismatch: %+v", usage.BuildCache)
	}

	// Check Totals
	expectedTotal := int64(150+30+70+60) * 1024 * 1024
	expectedRecl := int64(50+20+40+60) * 1024 * 1024
	if usage.TotalSize != expectedTotal {
		t.Errorf("expected total size %d, got %d", expectedTotal, usage.TotalSize)
	}
	if usage.TotalRecl != expectedRecl {
		t.Errorf("expected reclaimable %d, got %d", expectedRecl, usage.TotalRecl)
	}
}

func TestGetDockerDiskUsageSafeExecution(t *testing.T) {
	// Function must safely execute without panics
	_, _ = GetDockerDiskUsage()
}

func TestFormatPruneCLIOutput(t *testing.T) {
	outputWithReclaimed := "Deleted Containers:\nabc123\nTotal reclaimed space: 14.5MB"
	res := formatPruneCLIOutput(outputWithReclaimed, "контейнеров")
	if !strings.Contains(res, "14.5MB") || !strings.Contains(res, "контейнеров") {
		t.Errorf("unexpected output: %s", res)
	}

	outputWithoutReclaimed := "Deleted Images: 0"
	res2 := formatPruneCLIOutput(outputWithoutReclaimed, "образов")
	if !strings.Contains(res2, "успешно выполнена") {
		t.Errorf("unexpected fallback output: %s", res2)
	}
}
