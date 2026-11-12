package tui

import (
	"fmt"
	"time"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/monitor"
	tea "github.com/charmbracelet/bubbletea"
)

// TimeFormat is the standard format used for dashboard timestamps.
const TimeFormat = "15:04:05"

// Prune options constants.
const (
	PruneOptionContainers = 1
	PruneOptionImages     = 2
	PruneOptionVolumes    = 3
	PruneOptionAll        = 4
)

func fetchAllData() dataFetchedMsg {
	dockerInfo := monitor.GetDockerSummary()
	return dataFetchedMsg{
		sysInfo:    monitor.GetSystemInfo(),
		cpuInfo:    monitor.GetCPUInfo(),
		memInfo:    monitor.GetMemoryInfo(),
		dockerInfo: dockerInfo,
		compose:    monitor.GetComposeProjects(dockerInfo.Containers),
		swarm:      monitor.GetSwarmSummary(),
		timestamp:  time.Now().Format(TimeFormat),
	}
}

func fetchDataCmd() tea.Cmd {
	return func() tea.Msg {
		return fetchAllData()
	}
}

func makeActionCmd(action func() error, errPrefix, successMsg string) tea.Cmd {
	return func() tea.Msg {
		if err := action(); err != nil {
			return actionFinishedMsg{err: fmt.Errorf("%s%w", errPrefix, err)}
		}
		return actionFinishedMsg{message: successMsg, refresh: true}
	}
}

func startContainerCmd(id, name string, dict i18n.Dictionary) tea.Cmd {
	return makeActionCmd(
		func() error { return monitor.StartContainer(id) },
		dict.ToastErrorStarting,
		fmt.Sprintf(dict.ToastContainerStarted, name),
	)
}

func stopContainerCmd(id, name string, dict i18n.Dictionary) tea.Cmd {
	return makeActionCmd(
		func() error { return monitor.StopContainer(id) },
		dict.ToastErrorStopping,
		fmt.Sprintf(dict.ToastContainerStopped, name),
	)
}

func restartContainerCmd(id, name string, dict i18n.Dictionary) tea.Cmd {
	return makeActionCmd(
		func() error { return monitor.RestartContainer(id) },
		dict.ToastErrorRestartingCont,
		fmt.Sprintf(dict.ToastContainerRest, name),
	)
}

func killContainerCmd(id, name string, dict i18n.Dictionary) tea.Cmd {
	return makeActionCmd(
		func() error { return monitor.KillContainer(id) },
		dict.ToastErrorKillingCont,
		fmt.Sprintf(dict.ToastContainerKilled, name),
	)
}

func removeContainerCmd(id, name string, dict i18n.Dictionary) tea.Cmd {
	return makeActionCmd(
		func() error { return monitor.RemoveContainer(id) },
		dict.ToastErrorRemovingCont,
		fmt.Sprintf(dict.ToastContainerRemoved, name),
	)
}

func startComposeCmd(name string, containers []monitor.ContainerItem, dict i18n.Dictionary) tea.Cmd {
	return makeActionCmd(
		func() error { return monitor.StartComposeProject(name, containers) },
		dict.ToastStartingCompose,
		fmt.Sprintf(dict.ToastComposeStarted, name),
	)
}

func stopComposeCmd(name string, containers []monitor.ContainerItem, dict i18n.Dictionary) tea.Cmd {
	return makeActionCmd(
		func() error { return monitor.StopComposeProject(name, containers) },
		dict.ToastStoppingCompose,
		fmt.Sprintf(dict.ToastComposeStopped, name),
	)
}

func restartComposeCmd(name string, containers []monitor.ContainerItem, dict i18n.Dictionary) tea.Cmd {
	return makeActionCmd(
		func() error { return monitor.RestartComposeProject(name, containers) },
		dict.ToastErrorRestarting,
		fmt.Sprintf(dict.ToastComposeRestarted, name),
	)
}

func downComposeCmd(name string, containers []monitor.ContainerItem, dict i18n.Dictionary) tea.Cmd {
	return makeActionCmd(
		func() error { return monitor.DownComposeProject(name, containers) },
		dict.ToastErrorDowning,
		fmt.Sprintf(dict.ToastComposeDowned, name),
	)
}

func executePrune(option int) (string, error) {
	switch option {
	case PruneOptionContainers:
		return monitor.PruneContainers()
	case PruneOptionImages:
		return monitor.PruneImages()
	case PruneOptionVolumes:
		return monitor.PruneVolumes()
	case PruneOptionAll:
		return monitor.PruneAll()
	default:
		return "", fmt.Errorf("unknown prune option: %d", option)
	}
}

func pruneCmd(option int, dict i18n.Dictionary) tea.Cmd {
	return func() tea.Msg {
		res, err := executePrune(option)
		if err != nil {
			return actionFinishedMsg{err: fmt.Errorf("%s%w", dict.ToastPruneError, err)}
		}
		return actionFinishedMsg{message: res, refresh: true}
	}
}
