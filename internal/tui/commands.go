package tui

import (
	"fmt"
	"time"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/monitor"
	tea "github.com/charmbracelet/bubbletea"
)

type actionFinishedMsg struct {
	message string
	err     error
	refresh bool
}

type dataFetchedMsg struct {
	sysInfo    monitor.SystemInfo
	cpuInfo    monitor.CPUInfo
	memInfo    monitor.MemoryInfo
	disksInfo  []monitor.PartitionInfo
	netInfo    monitor.NetworkInfo
	topProcs   []monitor.ProcessInfo
	dockerInfo monitor.DockerSummary
	compose    []monitor.ComposeProject
	swarm      monitor.SwarmSummary
	timestamp  string
}

func fetchDataCmd() tea.Cmd {
	return func() tea.Msg {
		dockerInfo := monitor.GetDockerSummary()
		return dataFetchedMsg{
			sysInfo:    monitor.GetSystemInfo(),
			cpuInfo:    monitor.GetCPUInfo(),
			memInfo:    monitor.GetMemoryInfo(),
			disksInfo:  monitor.GetDisksInfo(),
			netInfo:    monitor.GetNetworkInfo(),
			topProcs:   monitor.GetTopProcesses(5),
			dockerInfo: dockerInfo,
			compose:    monitor.GetComposeProjects(dockerInfo.Containers),
			swarm:      monitor.GetSwarmSummary(),
			timestamp:  time.Now().Format("15:04:05"),
		}
	}
}

func startContainerCmd(id, name string, dict i18n.Dictionary) tea.Cmd {
	return func() tea.Msg {
		if err := monitor.StartContainer(id); err != nil {
			return actionFinishedMsg{err: fmt.Errorf("%s%w", dict.ToastErrorStarting, err)}
		}
		return actionFinishedMsg{message: fmt.Sprintf(dict.ToastContainerStarted, name), refresh: true}
	}
}

func stopContainerCmd(id, name string, dict i18n.Dictionary) tea.Cmd {
	return func() tea.Msg {
		if err := monitor.StopContainer(id); err != nil {
			return actionFinishedMsg{err: fmt.Errorf("%s%w", dict.ToastErrorStopping, err)}
		}
		return actionFinishedMsg{message: fmt.Sprintf(dict.ToastContainerStopped, name), refresh: true}
	}
}

func restartContainerCmd(id, name string, dict i18n.Dictionary) tea.Cmd {
	return func() tea.Msg {
		if err := monitor.RestartContainer(id); err != nil {
			return actionFinishedMsg{err: fmt.Errorf("%s%w", dict.ToastErrorRestartingCont, err)}
		}
		return actionFinishedMsg{message: fmt.Sprintf(dict.ToastContainerRest, name), refresh: true}
	}
}

func startComposeCmd(name string, containers []monitor.ContainerItem, dict i18n.Dictionary) tea.Cmd {
	return func() tea.Msg {
		if err := monitor.StartComposeProject(name, containers); err != nil {
			return actionFinishedMsg{err: fmt.Errorf("%s%w", dict.ToastStartingCompose, err)}
		}
		return actionFinishedMsg{message: fmt.Sprintf(dict.ToastComposeStarted, name), refresh: true}
	}
}

func stopComposeCmd(name string, containers []monitor.ContainerItem, dict i18n.Dictionary) tea.Cmd {
	return func() tea.Msg {
		if err := monitor.StopComposeProject(name, containers); err != nil {
			return actionFinishedMsg{err: fmt.Errorf("%s%w", dict.ToastStoppingCompose, err)}
		}
		return actionFinishedMsg{message: fmt.Sprintf(dict.ToastComposeStopped, name), refresh: true}
	}
}

func restartComposeCmd(name string, containers []monitor.ContainerItem, dict i18n.Dictionary) tea.Cmd {
	return func() tea.Msg {
		if err := monitor.RestartComposeProject(name, containers); err != nil {
			return actionFinishedMsg{err: fmt.Errorf("%s%w", dict.ToastErrorRestarting, err)}
		}
		return actionFinishedMsg{message: fmt.Sprintf(dict.ToastComposeRestarted, name), refresh: true}
	}
}

func downComposeCmd(name string, containers []monitor.ContainerItem, dict i18n.Dictionary) tea.Cmd {
	return func() tea.Msg {
		if err := monitor.DownComposeProject(name, containers); err != nil {
			return actionFinishedMsg{err: fmt.Errorf("%s%w", dict.ToastErrorDowning, err)}
		}
		return actionFinishedMsg{message: fmt.Sprintf(dict.ToastComposeDowned, name), refresh: true}
	}
}

func pruneCmd(option int, dict i18n.Dictionary) tea.Cmd {
	return func() tea.Msg {
		var res string
		var err error
		switch option {
		case 1:
			res, err = monitor.PruneContainers()
		case 2:
			res, err = monitor.PruneImages()
		case 3:
			res, err = monitor.PruneVolumes()
		case 4:
			res, err = monitor.PruneAll()
		}
		if err != nil {
			return actionFinishedMsg{err: fmt.Errorf("%s%w", dict.ToastPruneError, err)}
		}
		return actionFinishedMsg{message: res, refresh: true}
	}
}
