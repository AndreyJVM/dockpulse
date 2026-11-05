package views

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/monitor"
	"github.com/AndreyJVM/dockpulse/internal/tui/components"
	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
)

// HostContext contains the data needed to render the Host Overview tab.
type HostContext struct {
	SysInfo               monitor.SystemInfo
	CPUInfo               monitor.CPUInfo
	MemInfo               monitor.MemoryInfo
	DockerInfo            monitor.DockerSummary
	CPUHistory            []float64
	MemHistory            []float64
	FormattedUptime       string
	FormattedDockerStatus string
	I18n                  *i18n.Dictionary
}

// RenderHost renders the Host Overview tab.
func RenderHost(ctx HostContext) string {
	var s strings.Builder

	// 1. Host System Information
	s.WriteString(theme.TitleStyle.Render(ctx.I18n.HostSysInfoTitle) + "\n")
	s.WriteString(fmt.Sprintf("  %s %s    %s %s\n",
		theme.DimStyle.Render(ctx.I18n.HostLabel), theme.BoldStyle.Render(ctx.SysInfo.Hostname),
		theme.DimStyle.Render(ctx.I18n.UptimeLabel), theme.GreenStyle.Render(ctx.FormattedUptime),
	))
	s.WriteString(fmt.Sprintf("  %s %s (%s %s)\n\n",
		theme.DimStyle.Render(ctx.I18n.OSLabel), ctx.SysInfo.OS, ctx.I18n.KernelLabel, ctx.SysInfo.Kernel,
	))

	// 2. Host Resources (CPU & RAM)
	s.WriteString(theme.TitleStyle.Render(ctx.I18n.HostPerfTitle) + "\n")
	s.WriteString(fmt.Sprintf("  %s %s  %s\n",
		fmt.Sprintf(ctx.I18n.CPUCoresLabel, ctx.CPUInfo.Cores), components.FormatBar(ctx.CPUInfo.Percent, 22), components.RenderSparkline(ctx.CPUHistory),
	))
	s.WriteString(fmt.Sprintf("  %s %s  %s\n\n",
		fmt.Sprintf(ctx.I18n.RAMLabel, ctx.MemInfo.UsedGB, ctx.MemInfo.TotalGB), components.FormatBar(ctx.MemInfo.Percent, 22), components.RenderSparkline(ctx.MemHistory),
	))

	// 3. Docker Service Status
	s.WriteString(theme.TitleStyle.Render(ctx.I18n.DockerStatusTitle) + "\n")
	s.WriteString(fmt.Sprintf("  %s %s    %s\n",
		ctx.I18n.DaemonStatus, ctx.FormattedDockerStatus,
		theme.BoldStyle.Render(fmt.Sprintf(ctx.I18n.ContainersSummary, ctx.DockerInfo.RunningContainers, ctx.DockerInfo.StoppedContainers)),
	))
	s.WriteString(fmt.Sprintf("  %s %d    %s %s\n",
		theme.DimStyle.Render(ctx.I18n.ImagesCountLabel), ctx.DockerInfo.ImagesCount,
		theme.DimStyle.Render(ctx.I18n.VersionLabel), ctx.DockerInfo.Version,
	))

	return s.String()
}
