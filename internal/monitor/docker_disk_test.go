package monitor

import (
	"context"
	"errors"
	"net/http"
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

func TestParseCLIInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"42", 42},
		{"0", 0},
		{" 100 ", 100},
		{"15 (reclaimable)", 15},
		{"invalid", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseCLIInt(tt.input)
			if got != tt.expected {
				t.Errorf("parseCLIInt(%q) = %d; want %d", tt.input, got, tt.expected)
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

func TestGetDockerDiskUsage_Socket(t *testing.T) {
	origFinder := dockerSocketFinder
	origFactory := socketClientFactory
	defer func() {
		dockerSocketFinder = origFinder
		socketClientFactory = origFactory
	}()

	dockerSocketFinder = func() string { return "/tmp/mock-docker.sock" }
	socketClientFactory = func(_ string) *http.Client {
		return newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "/system/df") {
				var raw systemDFRaw
				raw.Images = []struct {
					Size       int64 `json:"Size"`
					SharedSize int64 `json:"SharedSize"`
					Containers int64 `json:"Containers"`
				}{
					{Size: 200 * 1024 * 1024, Containers: 0},
				}
				return mockJSONResponse(http.StatusOK, raw), nil
			}
			return mockStringResponse(http.StatusNotFound, "not found"), nil
		})
	}

	usage, err := GetDockerDiskUsage()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if usage.Images.TotalCount != 1 || usage.Images.Reclaimable != 200*1024*1024 {
		t.Errorf("unexpected usage data: %+v", usage.Images)
	}
}

func TestGetDockerDiskUsage_CLIFallback(t *testing.T) {
	origFinder := dockerSocketFinder
	origRunner := dockerCLICommandRunner
	defer func() {
		dockerSocketFinder = origFinder
		dockerCLICommandRunner = origRunner
	}()

	dockerSocketFinder = func() string { return "" }
	dockerCLICommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
		lines := []string{
			`{"Type":"Images","TotalCount":"5","Active":"2","Size":"100MB","Reclaimable":"20MB"}`,
			`{"Type":"Containers","TotalCount":"8","Active":"3","Size":"50MB","Reclaimable":"10MB"}`,
			`{"Type":"Local Volumes","TotalCount":"4","Active":"1","Size":"30MB","Reclaimable":"20MB"}`,
			`{"Type":"Build Cache","TotalCount":"6","Active":"2","Size":"40MB","Reclaimable":"10MB"}`,
		}
		return []byte(strings.Join(lines, "\n")), nil
	}

	usage, err := GetDockerDiskUsage()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if usage.Images.TotalCount != 5 || usage.Images.ActiveCount != 2 {
		t.Errorf("unexpected Images: %+v", usage.Images)
	}
	if usage.Containers.TotalCount != 8 || usage.Containers.ActiveCount != 3 {
		t.Errorf("unexpected Containers: %+v", usage.Containers)
	}
	if usage.Volumes.TotalCount != 4 || usage.Volumes.ActiveCount != 1 {
		t.Errorf("unexpected Volumes: %+v", usage.Volumes)
	}
	if usage.BuildCache.TotalCount != 6 || usage.BuildCache.ActiveCount != 2 {
		t.Errorf("unexpected BuildCache: %+v", usage.BuildCache)
	}
}

func TestGetDockerDiskUsage_Error(t *testing.T) {
	origFinder := dockerSocketFinder
	origRunner := dockerCLICommandRunner
	defer func() {
		dockerSocketFinder = origFinder
		dockerCLICommandRunner = origRunner
	}()

	dockerSocketFinder = func() string { return "" }
	dockerCLICommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
		return nil, errors.New("command failed")
	}

	_, err := GetDockerDiskUsage()
	if err == nil {
		t.Errorf("expected error when both socket and CLI fail")
	}
}
