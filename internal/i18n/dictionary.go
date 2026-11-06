package i18n

import "strings"

// Dictionary содержит все локализуемые строки пользовательского интерфейса.
type Dictionary struct {
	// Вкладки
	TabHost    string
	TabDocker  string
	TabCompose string
	TabSwarm   string

	// Статусы компонентов
	StatusRunning      string
	StatusStopped      string
	StatusNotRunning   string
	StatusNotInstalled string
	StatusPartial      string
	StatusActive       string
	StatusInactive     string

	// Форматы аптайма
	UptimeDaysFormat  string
	UptimeHoursFormat string

	// Подвал и навигация
	FooterTabs          string
	FooterRefresh       string
	FooterHelp          string
	FooterQuit          string
	FooterSelect        string
	FooterSearch        string
	FooterLogs          string
	FooterInfo          string
	FooterShell         string
	FooterStartStop     string
	FooterStart         string
	FooterStop          string
	FooterRestart       string
	FooterKill          string
	FooterRemove        string
	FooterPrune         string
	FooterDown          string
	FooterSort          string
	FooterApplyFilter   string
	FooterCancel        string
	FooterResetFilter   string
	FooterConfirm       string
	FooterClose         string
	FooterScroll        string
	FooterSelectStack   string
	FooterSelectService string
	FooterServiceLogs   string
	FooterSelectAction  string
	FooterConfirmPrune  string

	// Вкладка 1: Host
	HostSysInfoTitle  string
	HostLabel         string
	UptimeLabel       string
	OSLabel           string
	KernelLabel       string
	HostPerfTitle     string
	CPUCoresLabel     string
	RAMLabel          string
	DockerStatusTitle string
	DaemonStatus      string
	ContainersSummary string

	// Вкладка 2: Docker
	DockerTabTitle         string
	ServiceLabel           string
	ImagesCountLabel       string
	VersionLabel           string
	ContainersNotRunning   string
	DockerDaemonDown       string
	FilterPrompt           string
	FilterHint             string
	FilterActiveBadge      string
	FilterResultsCount     string
	NoContainersFound      string
	PressEscToClear        string
	ColID                  string
	ColName                string
	ColImage               string
	ColCPU                 string
	ColMem                 string
	ColState               string
	ColHealth              string
	ColPorts               string
	ColStatus              string
	SelectedContainerDetails string
	DetailPortsLabel       string
	DetailMountsLabel      string
	DetailNone             string
	DetailComposeLabel     string
	SortBadge              string
	SortModeDefault        string
	SortModeCPU            string
	SortModeMem            string
	SortModeName           string
	SortModeStatus         string

	// Вкладка 3: Compose
	ComposeTabTitle   string
	NoComposeProjects string
	ComposeServicesUp string

	// Вкладка 4: Swarm
	SwarmTabTitle     string
	SwarmInactive     string
	SwarmInitHint     string
	SwarmNodesCount   string
	SwarmManagers     string
	SwarmNodesTitle   string
	ColNodeID         string
	ColHostname       string
	ColRole           string
	ColAvailability   string
	ColTasks          string
	ColLeader         string
	LeaderBadge       string
	SwarmSvcsTitle    string
	ColServiceID      string
	ColReplicas       string

	// Модальное окно: Логи
	LogsTitle      string
	NoLogsOutput   string
	LogsRangeCount string

	// Модальное окно: Инспекция
	InspectTitle  string
	ImageLabel    string
	HealthLabel   string
	StatusLabel   string
	MemLabel      string
	CreatedLabel  string
	IPAddrLabel   string
	GatewayLabel  string
	CommandLabel  string
	MountsTitle   string

	// Модальное окно: Prune
	PruneTitle         string
	CurrentDiskUsage   string
	TypeCol            string
	SizeCol            string
	ActiveTotalCol     string
	ReclaimableCol     string
	TypeImages         string
	TypeContainers     string
	TypeVolumes        string
	TypeBuildCache     string
	TypeTotal          string
	PruneChooseAction  string
	PruneOpt1          string
	PruneOpt2          string
	PruneOpt3          string
	PruneOpt4          string
	PruneWarnTitle     string
	PruneWarnAction1   string
	PruneWarnAction2   string
	PruneWarnAction3   string
	PruneWarnAction4   string
	PruneWarnIrrev     string

	// Модальное окно: Compose Down
	ComposeDownTitle   string
	StackLabel         string
	WarningLabel       string
	ComposeDownNotice  string
	ComposeDownConfirm string

	// Модальное окно: Remove Container
	RemoveContainerTitle   string
	ContainerLabel         string
	RemoveContainerNotice  string
	RemoveContainerConfirm string

	// Модальное окно: Справка
	HelpTitle     string
	HelpKeyTabs   string
	HelpKeyPrev   string
	HelpKeyNav    string
	HelpKeySearch string
	HelpKeyLogs   string
	HelpKeyInfo   string
	HelpKeyShell  string
	HelpKeySS     string
	HelpKeyStart  string
	HelpKeyRest   string
	HelpKeyKill   string
	HelpKeyRemove string
	HelpKeyDown   string
	HelpKeyPrune  string
	HelpKeySort   string
	HelpKeyRef    string
	HelpKeyHelp   string
	HelpKeyQuit   string
	HelpAutoInt   string

	// Уведомления (Toasts)
	ToastDataRefreshed       string
	ToastRefreshing          string
	ToastStoppingContainer   string
	ToastContainerStopped    string
	ToastStartingContainer   string
	ToastContainerStarted    string
	ToastRestartingCont      string
	ToastContainerRest       string
	ToastCannotConnect       string
	ToastDockerNotFound      string
	ToastShellEnded          string
	ToastShellEndedCode      string
	ToastShellError          string
	ToastDowningCompose      string
	ToastComposeDowned       string
	ToastErrorDowning        string
	ToastStoppingCompose     string
	ToastComposeStopped      string
	ToastStartingCompose     string
	ToastComposeStarted      string
	ToastRestartingCompose   string
	ToastComposeRestarted    string
	ToastErrorRestarting     string
	ToastPruning             string
	ToastPruneError          string
	ToastInspectError        string
	ToastLogsError           string
	ToastErrorStarting       string
	ToastErrorStopping       string
	ToastErrorRestartingCont string
	ToastKillingCont         string
	ToastContainerKilled     string
	ToastErrorKillingCont    string
	ToastRemovingCont        string
	ToastContainerRemoved    string
	ToastErrorRemovingCont   string
}

// Ru — русская локализация.
var Ru = Dictionary{
	TabHost:    "1: Host Overview",
	TabDocker:  "2: Docker Containers",
	TabCompose: "3: Compose Stacks",
	TabSwarm:   "4: Swarm Cluster",

	StatusRunning:      "Запущен",
	StatusStopped:      "Остановлен",
	StatusNotRunning:   "Не запущен",
	StatusNotInstalled: "Не установлен",
	StatusPartial:      "Частично",
	StatusActive:       "Активен",
	StatusInactive:     "Не активен",

	UptimeDaysFormat:  "%d дн %02d:%02d",
	UptimeHoursFormat: "%02d:%02d",

	FooterTabs:          "Вкладки",
	FooterRefresh:       "Обновить",
	FooterHelp:          "Справка",
	FooterQuit:          "Выход",
	FooterSelect:        "Выбрать",
	FooterSearch:        "Поиск",
	FooterLogs:          "Логи",
	FooterInfo:          "Инфо",
	FooterShell:         "Shell",
	FooterStartStop:     "Старт/Стоп",
	FooterStart:         "Старт",
	FooterStop:          "Стоп",
	FooterRestart:       "Рестарт",
	FooterKill:          "Kill",
	FooterRemove:        "Удалить",
	FooterPrune:         "Очистка",
	FooterDown:          "Down",
	FooterSort:          "Сортировка",
	FooterApplyFilter:   "Применить фильтр",
	FooterCancel:        "Отмена",
	FooterResetFilter:   "Сбросить фильтр",
	FooterConfirm:       "Подтвердить",
	FooterClose:         "Закрыть окно",
	FooterScroll:        "Прокрутка",
	FooterSelectStack:   "Выбрать стек",
	FooterSelectService: "Выбрать службу",
	FooterServiceLogs:   "Логи службы",
	FooterSelectAction:  "Выбрать действие",
	FooterConfirmPrune:  "Подтвердить очистку",

	HostSysInfoTitle:  "СИСТЕМНАЯ ИНФОРМАЦИЯ",
	HostLabel:         "Хост:",
	UptimeLabel:       "Uptime:",
	OSLabel:           "ОС:",
	KernelLabel:       "Kernel:",
	HostPerfTitle:     "ПРОИЗВОДИТЕЛЬНОСТЬ ХОСТА",
	CPUCoresLabel:     "CPU (%d cores):",
	RAMLabel:          "RAM (%.1f/%.1f GB):",
	DockerStatusTitle: "DOCKER СТАТУС",
	DaemonStatus:      "Статус:",
	ContainersSummary: "Контейнеры: %d активно, %d остановлено",

	DockerTabTitle:       "DOCKER КОНТЕЙНЕРЫ (ИНТЕРАКТИВНОЕ УПРАВЛЕНИЕ)",
	ServiceLabel:         "Служба:",
	ImagesCountLabel:     "Образов:",
	VersionLabel:         "Версия:",
	ContainersNotRunning: "Контейнеры не запущены.",
	DockerDaemonDown:     "Демон Docker не запущен или недоступен.",
	FilterPrompt:         "[Поиск: %s█]",
	FilterHint:           "(Enter: применить, Esc: отмена)",
	FilterActiveBadge:    "Фильтр: %s",
	FilterResultsCount:   "(Найдено: %d из %d, Esc: сбросить)",
	NoContainersFound:    "По запросу %q контейнеры не найдены.",
	PressEscToClear:      "Нажмите Esc для сброса поискового фильтра.",
	ColID:                "ID",
	ColName:              "NAME",
	ColImage:             "IMAGE",
	ColCPU:               "CPU%",
	ColMem:               "MEM USAGE",
	ColState:             "STATE",
	ColHealth:            "HEALTH",
	ColPorts:             "PORTS",
	ColStatus:            "СТАТУС",
	SelectedContainerDetails: "ДЕТАЛИ ВЫБРАННОГО КОНТЕЙНЕРА",
	DetailPortsLabel:       "Порты:",
	DetailMountsLabel:      "Тома / Mounts:",
	DetailNone:             "нет",
	DetailComposeLabel:     "Compose:",
	SortBadge:              "Сортировка: %s",
	SortModeDefault:        "По умолчанию",
	SortModeCPU:            "CPU ↓",
	SortModeMem:            "Память ↓",
	SortModeName:           "Имя ↑",
	SortModeStatus:         "Статус",

	ComposeTabTitle:   "DOCKER COMPOSE ПРОЕКТЫ (%d)",
	NoComposeProjects: "Проекты Docker Compose не обнаружены (нет меток com.docker.compose.project).",
	ComposeServicesUp: "(Сервисов поднято: %d/%d)",

	SwarmTabTitle:   "DOCKER SWARM КЛАСТЕР",
	SwarmInactive:   "Хост не входит в активный кластер Swarm.",
	SwarmInitHint:   "Для создания кластера выполните: docker swarm init",
	SwarmNodesCount: "Нод в кластере:",
	SwarmManagers:   "Менеджеров:",
	SwarmNodesTitle: "УЗЛЫ КЛАСТЕРА (NODES):",
	ColNodeID:       "NODE ID",
	ColHostname:     "HOSTNAME",
	ColRole:         "ROLE",
	ColAvailability: "AVAILABILITY",
	ColTasks:        "TASKS",
	ColLeader:       "LEADER",
	LeaderBadge:     "[Leader]",
	SwarmSvcsTitle:  "СЛУЖБЫ SWARM (SERVICES):",
	ColServiceID:    "SERVICE ID",
	ColReplicas:     "REPLICAS",

	LogsTitle:      "ЛОГИ: %s (ID: %s)",
	NoLogsOutput:   "Логи отсутствуют или служба не выводила сообщений.",
	LogsRangeCount: "[Строки %d-%d из %d] (прокрутка ↑/↓/PgUp/PgDn)",

	InspectTitle: "ИНСПЕКЦИЯ КОНТЕЙНЕРА: %s",
	ImageLabel:   "Образ:",
	HealthLabel:  "Здоровье:",
	StatusLabel:  "Статус:",
	MemLabel:     "Память:",
	CreatedLabel: "Создан:",
	IPAddrLabel:  "IP-адрес:",
	GatewayLabel: "Шлюз:",
	CommandLabel: "Команда:",
	MountsTitle:  "Тома и монтирования (Mounts):",

	PruneTitle:        "ОЧИСТКА РЕСУРСОВ DOCKER (SYSTEM PRUNE)",
	CurrentDiskUsage:  "ТЕКУЩЕЕ ИСПОЛЬЗОВАНИЕ ДИСКА:",
	TypeCol:           "ТИП",
	SizeCol:           "ОБЪЕМ",
	ActiveTotalCol:    "АКТИВНО/ВСЕГО",
	ReclaimableCol:    "ОСВОБОЖДАЕМО",
	TypeImages:        "Образы",
	TypeContainers:    "Контейнеры",
	TypeVolumes:       "Тома",
	TypeBuildCache:    "Кэш сборки",
	TypeTotal:         "Всего",
	PruneChooseAction: "ВЫБЕРИТЕ ДЕЙСТВИЕ ДЛЯ ОЧИСТКИ:",
	PruneOpt1:         "Очистить остановленные контейнеры (Prune stopped containers)",
	PruneOpt2:         "Очистить висячие образы без тегов (Prune dangling images)",
	PruneOpt3:         "Очистить неиспользуемые тома данных (Prune unused volumes)",
	PruneOpt4:         "Полная очистка всех неиспользуемых ресурсов (Prune all)",
	PruneWarnTitle:    "[ВНИМАНИЕ] ",
	PruneWarnAction1:  "Удаление всех остановленных контейнеров",
	PruneWarnAction2:  "Удаление всех висячих образов (dangling images)",
	PruneWarnAction3:  "Удаление всех неиспользуемых томов (volumes)",
	PruneWarnAction4:  "ПОЛНАЯ очистка: остановленные контейнеры, образы, тома и кэш сборки",
	PruneWarnIrrev:    "Это действие необратимо. Подтвердить выполнение?",

	ComposeDownTitle:   "ПОДТВЕРЖДЕНИЕ СВОРАЧИВАНИЯ COMPOSE СТЕКА",
	StackLabel:         "Стек: %s",
	WarningLabel:       "Внимание:",
	ComposeDownNotice:  "Остановка и удаление всех контейнеров проекта (docker compose down)",
	ComposeDownConfirm: "Вы уверены, что хотите выполнить down для проекта %s?",

	RemoveContainerTitle:   "ПОДТВЕРЖДЕНИЕ УДАЛЕНИЯ КОНТЕЙНЕРА",
	ContainerLabel:         "Контейнер: %s",
	RemoveContainerNotice:  "Контейнер и его временные данные будут безвозвратно удалены (docker rm -f).",
	RemoveContainerConfirm: "Вы уверены, что хотите принудительно удалить контейнер %s?",

	HelpTitle:     "ГОРЯЧИЕ КЛАВИШИ:",
	HelpKeyTabs:   "Переключение вкладок (1: Host, 2: Docker, 3: Compose, 4: Swarm)",
	HelpKeyPrev:   "След. / пред. вкладка",
	HelpKeyNav:    "Перемещение курсора (контейнеры, стеки Compose, службы Swarm)",
	HelpKeySearch: "Поиск и фильтрация контейнеров по имени, образу или ID",
	HelpKeyLogs:   "Просмотр логов контейнера, стека Compose или службы Swarm",
	HelpKeyInfo:   "Подробная инспекция выбранного контейнера (IP, тома, сеть)",
	HelpKeyShell:  "Интерактивный shell в контейнере (bash / sh)",
	HelpKeySS:     "Запуск / остановка контейнера, остановка Compose проекта",
	HelpKeyStart:  "Запуск остановленного проекта Compose (Start)",
	HelpKeyRest:   "Перезапуск (restart) контейнера или проекта Compose",
	HelpKeyKill:   "Принудительная остановка контейнера (SIGKILL)",
	HelpKeyRemove: "Принудительное удаление контейнера (Remove)",
	HelpKeyDown:   "Сворачивание и удаление проекта Compose (Down)",
	HelpKeyPrune:  "Очистка неиспользуемых ресурсов Docker (Prune)",
	HelpKeySort:   "Сортировка контейнеров (CPU / Память / Имя / Статус)",
	HelpKeyRef:    "Принудительное обновление всех данных",
	HelpKeyHelp:   "Показать / скрыть эту справку",
	HelpKeyQuit:   "Выход из приложения",
	HelpAutoInt:   "Интервал автообновления: %s",

	ToastDataRefreshed:       "Данные успешно обновлены",
	ToastRefreshing:          "Обновление данных...",
	ToastStoppingContainer:   "Остановка контейнера %s...",
	ToastContainerStopped:    "Контейнер %s остановлен",
	ToastStartingContainer:   "Запуск контейнера %s...",
	ToastContainerStarted:    "Контейнер %s запущен",
	ToastRestartingCont:      "Перезапуск %s...",
	ToastContainerRest:       "Контейнер %s успешно перезапущен",
	ToastCannotConnect:       "Нельзя подключиться: контейнер %s не запущен",
	ToastDockerNotFound:      "Ошибка: утилита docker не найдена в $PATH",
	ToastShellEnded:          "Сессия shell в %s завершена",
	ToastShellEndedCode:      "Сессия shell в %s завершена (код %d)",
	ToastShellError:          "Ошибка shell в %s: %v",
	ToastDowningCompose:      "Сворачивание Compose проекта %s (down)...",
	ToastComposeDowned:       "Проект %s успешно свернут (down)",
	ToastErrorDowning:        "Ошибка сворачивания проекта: ",
	ToastStoppingCompose:     "Остановка Compose проекта %s...",
	ToastComposeStopped:      "Проект %s успешно остановлен",
	ToastStartingCompose:     "Запуск Compose проекта %s...",
	ToastComposeStarted:      "Проект %s успешно запущен",
	ToastRestartingCompose:   "Перезапуск Compose проекта %s...",
	ToastComposeRestarted:    "Проект %s успешно перезапущен",
	ToastErrorRestarting:     "Ошибка перезапуска проекта: ",
	ToastPruning:             "Выполняется очистка ресурсов Docker...",
	ToastPruneError:          "Ошибка очистки: ",
	ToastInspectError:        "Ошибка inspect: ",
	ToastLogsError:           "Ошибка чтения логов: ",
	ToastErrorStarting:       "Ошибка запуска контейнера: ",
	ToastErrorStopping:       "Ошибка остановки контейнера: ",
	ToastErrorRestartingCont: "Ошибка перезапуска контейнера: ",
	ToastKillingCont:         "Принудительная остановка %s (kill)...",
	ToastContainerKilled:     "Контейнер %s принудительно остановлен (kill)",
	ToastErrorKillingCont:    "Ошибка принудительной остановки: ",
	ToastRemovingCont:        "Удаление контейнера %s...",
	ToastContainerRemoved:    "Контейнер %s успешно удален",
	ToastErrorRemovingCont:   "Ошибка удаления контейнера: ",
}

// En — английская локализация.
var En = Dictionary{
	TabHost:    "1: Host Overview",
	TabDocker:  "2: Docker Containers",
	TabCompose: "3: Compose Stacks",
	TabSwarm:   "4: Swarm Cluster",

	StatusRunning:      "Running",
	StatusStopped:      "Stopped",
	StatusNotRunning:   "Not running",
	StatusNotInstalled: "Not installed",
	StatusPartial:      "Partial",
	StatusActive:       "Active",
	StatusInactive:     "Inactive",

	UptimeDaysFormat:  "%dd %02d:%02d",
	UptimeHoursFormat: "%02d:%02d",

	FooterTabs:          "Tabs",
	FooterRefresh:       "Refresh",
	FooterHelp:          "Help",
	FooterQuit:          "Quit",
	FooterSelect:        "Select",
	FooterSearch:        "Search",
	FooterLogs:          "Logs",
	FooterInfo:          "Info",
	FooterShell:         "Shell",
	FooterStartStop:     "Start/Stop",
	FooterStart:         "Start",
	FooterStop:          "Stop",
	FooterRestart:       "Restart",
	FooterKill:          "Kill",
	FooterRemove:        "Remove",
	FooterPrune:         "Prune",
	FooterDown:          "Down",
	FooterSort:          "Sort",
	FooterApplyFilter:   "Apply filter",
	FooterCancel:        "Cancel",
	FooterResetFilter:   "Reset filter",
	FooterConfirm:       "Confirm",
	FooterClose:         "Close window",
	FooterScroll:        "Scroll",
	FooterSelectStack:   "Select stack",
	FooterSelectService: "Select service",
	FooterServiceLogs:   "Service logs",
	FooterSelectAction:  "Select action",
	FooterConfirmPrune:  "Confirm prune",

	HostSysInfoTitle:  "SYSTEM INFORMATION",
	HostLabel:         "Host:",
	UptimeLabel:       "Uptime:",
	OSLabel:           "OS:",
	KernelLabel:       "Kernel:",
	HostPerfTitle:     "HOST PERFORMANCE",
	CPUCoresLabel:     "CPU (%d cores):",
	RAMLabel:          "RAM (%.1f/%.1f GB):",
	DockerStatusTitle: "DOCKER STATUS",
	DaemonStatus:      "Status:",
	ContainersSummary: "Containers: %d active, %d stopped",

	DockerTabTitle:       "DOCKER CONTAINERS (INTERACTIVE MANAGEMENT)",
	ServiceLabel:         "Daemon:",
	ImagesCountLabel:     "Images:",
	VersionLabel:         "Version:",
	ContainersNotRunning: "No containers running.",
	DockerDaemonDown:     "Docker daemon is not running or unavailable.",
	FilterPrompt:         "[Search: %s█]",
	FilterHint:           "(Enter: apply, Esc: cancel)",
	FilterActiveBadge:    "Filter: %s",
	FilterResultsCount:   "(Found: %d of %d, Esc: reset)",
	NoContainersFound:    "No containers found matching %q.",
	PressEscToClear:      "Press Esc to clear search filter.",
	ColID:                "ID",
	ColName:              "NAME",
	ColImage:             "IMAGE",
	ColCPU:               "CPU%",
	ColMem:               "MEM USAGE",
	ColState:             "STATE",
	ColHealth:            "HEALTH",
	ColPorts:             "PORTS",
	ColStatus:            "STATUS",
	SelectedContainerDetails: "SELECTED CONTAINER DETAILS",
	DetailPortsLabel:       "Ports:",
	DetailMountsLabel:      "Volumes / Mounts:",
	DetailNone:             "none",
	DetailComposeLabel:     "Compose:",
	SortBadge:              "Sort: %s",
	SortModeDefault:        "Default",
	SortModeCPU:            "CPU ↓",
	SortModeMem:            "Memory ↓",
	SortModeName:           "Name ↑",
	SortModeStatus:         "Status",

	ComposeTabTitle:   "DOCKER COMPOSE PROJECTS (%d)",
	NoComposeProjects: "No Docker Compose projects detected (missing com.docker.compose.project labels).",
	ComposeServicesUp: "(Services up: %d/%d)",

	SwarmTabTitle:   "DOCKER SWARM CLUSTER",
	SwarmInactive:   "Host is not part of an active Swarm cluster.",
	SwarmInitHint:   "To initialize a cluster, run: docker swarm init",
	SwarmNodesCount: "Cluster nodes:",
	SwarmManagers:   "Managers:",
	SwarmNodesTitle: "CLUSTER NODES:",
	ColNodeID:       "NODE ID",
	ColHostname:     "HOSTNAME",
	ColRole:         "ROLE",
	ColAvailability: "AVAILABILITY",
	ColTasks:        "TASKS",
	ColLeader:       "LEADER",
	LeaderBadge:     "[Leader]",
	SwarmSvcsTitle:  "SWARM SERVICES:",
	ColServiceID:    "SERVICE ID",
	ColReplicas:     "REPLICAS",

	LogsTitle:      "LOGS: %s (ID: %s)",
	NoLogsOutput:   "No logs available or service produced no output.",
	LogsRangeCount: "[Lines %d-%d of %d] (scroll ↑/↓/PgUp/PgDn)",

	InspectTitle: "CONTAINER INSPECTION: %s",
	ImageLabel:   "Image:",
	HealthLabel:  "Health:",
	StatusLabel:  "Status:",
	MemLabel:     "Memory:",
	CreatedLabel: "Created:",
	IPAddrLabel:  "IP Address:",
	GatewayLabel: "Gateway:",
	CommandLabel: "Command:",
	MountsTitle:  "Mounts and Volumes:",

	PruneTitle:        "DOCKER RESOURCE PRUNE (SYSTEM PRUNE)",
	CurrentDiskUsage:  "CURRENT DISK USAGE:",
	TypeCol:           "TYPE",
	SizeCol:           "SIZE",
	ActiveTotalCol:    "ACTIVE/TOTAL",
	ReclaimableCol:    "RECLAIMABLE",
	TypeImages:        "Images",
	TypeContainers:    "Containers",
	TypeVolumes:       "Volumes",
	TypeBuildCache:    "Build Cache",
	TypeTotal:         "Total",
	PruneChooseAction: "CHOOSE ACTION TO PRUNE:",
	PruneOpt1:         "Prune stopped containers",
	PruneOpt2:         "Prune dangling images",
	PruneOpt3:         "Prune unused volumes",
	PruneOpt4:         "Prune all unused resources",
	PruneWarnTitle:    "[WARNING] ",
	PruneWarnAction1:  "Remove all stopped containers",
	PruneWarnAction2:  "Remove all dangling untagged images",
	PruneWarnAction3:  "Remove all unused data volumes",
	PruneWarnAction4:  "FULL prune: stopped containers, images, volumes, and build cache",
	PruneWarnIrrev:    "This action is irreversible. Confirm execution?",

	ComposeDownTitle:   "CONFIRM DOCKER COMPOSE DOWN",
	StackLabel:         "Stack: %s",
	WarningLabel:       "Warning:",
	ComposeDownNotice:  "Stopping and removing all project containers (docker compose down)",
	ComposeDownConfirm: "Are you sure you want to down project %s?",

	RemoveContainerTitle:   "CONFIRM CONTAINER REMOVAL",
	ContainerLabel:         "Container: %s",
	RemoveContainerNotice:  "The container and its ephemeral data will be permanently removed (docker rm -f).",
	RemoveContainerConfirm: "Are you sure you want to force remove container %s?",

	HelpTitle:     "HOTKEYS:",
	HelpKeyTabs:   "Switch tabs (1: Host, 2: Docker, 3: Compose, 4: Swarm)",
	HelpKeyPrev:   "Next / previous tab",
	HelpKeyNav:    "Move cursor (containers, Compose stacks, Swarm services)",
	HelpKeySearch: "Search and filter containers by name, image, or ID",
	HelpKeyLogs:   "View logs for container, Compose stack, or Swarm service",
	HelpKeyInfo:   "Detailed container inspection (IP, mounts, network)",
	HelpKeyShell:  "Interactive shell inside container (bash / sh)",
	HelpKeySS:     "Start / stop container, stop Compose project",
	HelpKeyStart:  "Start stopped Compose project (Start)",
	HelpKeyRest:   "Restart container or Compose project",
	HelpKeyKill:   "Force kill container (SIGKILL)",
	HelpKeyRemove: "Force remove container (Remove)",
	HelpKeyDown:   "Down and remove Compose project (Down)",
	HelpKeyPrune:  "Prune unused Docker resources (Prune)",
	HelpKeySort:   "Cycle container sorting (CPU / Memory / Name / Status)",
	HelpKeyRef:    "Force refresh all metrics",
	HelpKeyHelp:   "Show / hide this help box",
	HelpKeyQuit:   "Quit application",
	HelpAutoInt:   "Auto-refresh interval: %s",

	ToastDataRefreshed:       "Data updated successfully",
	ToastRefreshing:          "Refreshing data...",
	ToastStoppingContainer:   "Stopping container %s...",
	ToastContainerStopped:    "Container %s stopped",
	ToastStartingContainer:   "Starting container %s...",
	ToastContainerStarted:    "Container %s started",
	ToastRestartingCont:      "Restarting %s...",
	ToastContainerRest:       "Container %s restarted successfully",
	ToastCannotConnect:       "Cannot connect: container %s is not running",
	ToastDockerNotFound:      "Error: docker binary not found in $PATH",
	ToastShellEnded:          "Shell session in %s ended",
	ToastShellEndedCode:      "Shell session in %s ended (code %d)",
	ToastShellError:          "Shell error in %s: %v",
	ToastDowningCompose:      "Downing Compose project %s (down)...",
	ToastComposeDowned:       "Project %s downed successfully",
	ToastErrorDowning:        "Error downing project: ",
	ToastStoppingCompose:     "Stopping Compose project %s...",
	ToastComposeStopped:      "Project %s stopped successfully",
	ToastStartingCompose:     "Starting Compose project %s...",
	ToastComposeStarted:      "Project %s started successfully",
	ToastRestartingCompose:   "Restarting Compose project %s...",
	ToastComposeRestarted:    "Project %s restarted successfully",
	ToastErrorRestarting:     "Error restarting project: ",
	ToastPruning:             "Pruning Docker resources...",
	ToastPruneError:          "Prune error: ",
	ToastInspectError:        "Inspect error: ",
	ToastLogsError:           "Logs error: ",
	ToastErrorStarting:       "Error starting container: ",
	ToastErrorStopping:       "Error stopping container: ",
	ToastErrorRestartingCont: "Error restarting container: ",
	ToastKillingCont:         "Force killing %s...",
	ToastContainerKilled:     "Container %s force-killed successfully",
	ToastErrorKillingCont:    "Error killing container: ",
	ToastRemovingCont:        "Removing container %s...",
	ToastContainerRemoved:    "Container %s removed successfully",
	ToastErrorRemovingCont:   "Error removing container: ",
}

// Get возвращает словарь для указанного языка ("ru" или "en").
// По умолчанию (для неизвестных языков и "en") возвращается английский словарь.
func Get(lang string) Dictionary {
	if strings.EqualFold(lang, "ru") || strings.HasPrefix(strings.ToLower(lang), "ru") {
		return Ru
	}
	return En
}
