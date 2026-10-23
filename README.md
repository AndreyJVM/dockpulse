# System & Docker Monitor (Go Edition) ⚡

[![Release](https://img.shields.io/github/v/release/AndreyJVM/system-monitor?color=blue&logo=github)](https://github.com/AndreyJVM/system-monitor/releases)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)](https://github.com/AndreyJVM/system-monitor/releases)

Сверхбыстрый автономный терминальный TUI-монитор системных ресурсов хоста, **Docker**, **Docker Compose** и кластеров **Docker Swarm**, написанный на **Go (Golang)** на базе фреймворков **Bubble Tea** и **Lip Gloss**.

```text
┌─ System & Docker Monitor ───────────────────────────────────────────┐
│ [1: Host Overview]  [2: Docker Containers]  [3: Compose]  [4: Swarm]   │
├─────────────────────────────────────────────────────────────────────┤
│ 🖥️  СИСТЕМНАЯ ИНФОРМАЦИЯ                                             │
│   Хост: server           Uptime: 4 дн 12:40                          │
│   ОС:   Ubuntu 24.04     Kernel: 6.8.0-45-generic                   │
│                                                                     │
│ ⚡  ПРОИЗВОДИТЕЛЬНОСТЬ ХОСТА                                         │
│   CPU (8 cores):         [████████░░░░░░░░░░░░░░░░░░░░]  34.2%       │
│   RAM (6.2/16.0 GB):     [██████████░░░░░░░░░░░░░░░░░░]  38.8%       │
│                                                                     │
│ 🌐  СЕТЬ И ТРАФИК                                                   │
│   IP: 192.168.1.50       DNS: 1.1.1.1, 8.8.8.8                       │
│   Скорость: ↓ RX: 145.2 KB/s    ↑ TX: 24.8 KB/s                     │
│                                                                     │
│ 💾  ДИСКОВЫЕ РАЗДЕЛЫ                                                │
│   /                      [████████████░░░░░░░░]  58.4% (58.4/100 GB) │
│                                                                     │
│ 🔥  ТОП ПРОЦЕССОВ (ПО CPU)                                           │
│   PID      ИМЯ                      CPU %      RAM (MB)             │
│   ────────────────────────────────────────────────────────          │
│   1420     dockerd                  12.4%      184.2 MB             │
│   2891     postgres                 8.1%       312.0 MB             │
│   3102     node                     5.3%       94.6 MB              │
│                                                                     │
│ 🐳  DOCKER СТАТУС                                                   │
│   Статус: 🟢 Запущен     Контейнеры: 6 активно, 2 остановлено        │
├─────────────────────────────────────────────────────────────────────┤
│ [1-4 / Tab / ← →] Вкладки  |  [R] Обновить  |  [?] Справка  |  [Q] Выход │
└─────────────────────────────────────────────────────────────────────┘
```

---

## ✨ Возможности

- **🖥️ Состояние хоста в реальном времени**: Загрузка CPU с количеством ядер, память RAM, смонтированные дисковые разделы, скорость сети в реальном времени (RX/TX) и точный аптайм.
- **🔥 Топ процессов**: Автоматический вывод наиболее ресурсоемких процессов (PID, имя, CPU %, RSS RAM в МБ).
- **🐳 Docker Containers**: Прямой опрос unix-сокета (`/var/run/docker.sock`, rootless Docker, Docker Desktop, Podman) со временем ответа **<2 мс** без тяжелых SDK. Полный список контейнеров, статусы, образы и порты.
- **📦 Docker Compose**: Автоматическая группировка контейнеров по проектам и сервисам Compose, агрегированный статус стеков (активен, частичный, остановлен).
- **🌐 Docker Swarm**: Мониторинг кластера Swarm — роли нод (Manager/Worker, Leader), реплицированные службы и число реплик.
- **⚡ Интерактивный TUI**: Полноэкранный псевдографический интерфейс на Charm Bubble Tea & Lip Gloss с цветными прогресс-барами, навигацией по вкладкам (`1-4`, `Tab`, стрелки), всплывающей справкой (`?`), обновлением по `R` и авто-тиком.
- **🚀 Нулевые зависимости**: Скомпилированный статический бинарный файл без внешних библиотек и рантаймов (<12 МБ). Работает на любом дистрибутиве Linux, macOS (Apple Silicon и Intel) и Windows.

---

## 🚀 Установка и быстрый запуск

### Вариант 1. Установка готового `.deb` пакета (Debian / Ubuntu / Linux Mint):
```bash
# Для x86_64 / amd64:
wget https://github.com/AndreyJVM/system-monitor/releases/download/v0.3.0/system-monitor_0.3.0_amd64.deb
sudo dpkg -i system-monitor_0.3.0_amd64.deb

# Для ARM64 (Raspberry Pi, ARM-серверы):
wget https://github.com/AndreyJVM/system-monitor/releases/download/v0.3.0/system-monitor_0.3.0_arm64.deb
sudo dpkg -i system-monitor_0.3.0_arm64.deb

# Запуск из любого терминала
system-monitor
```

### Вариант 2. Запуск готового бинарника (Linux / macOS):
Скачайте архив со страницы [Releases](https://github.com/AndreyJVM/system-monitor/releases):
```bash
# Linux amd64
tar -xzf system-monitor_v0.3.0_linux_amd64.tar.gz
./system-monitor

# macOS (Apple Silicon)
tar -xzf system-monitor_v0.3.0_darwin_arm64.tar.gz
./system-monitor
```

### Вариант 3. Сборка из исходников (Go 1.21+):
```bash
git clone https://github.com/AndreyJVM/system-monitor.git
cd system-monitor

# Сборка бинарника
make build

# Запуск
./system-monitor
```

---

## ⌨️ Управление в TUI:

| Клавиша | Действие |
|---|---|
| `1` | Вкладка **Host Overview**: CPU, RAM, диски, сеть RX/TX, топ процессов, сводка Docker |
| `2` | Вкладка **Docker Containers**: Список контейнеров (ID, Name, Image, State, Ports) |
| `3` | Вкладка **Compose Stacks**: Стеки и сервисы Docker Compose |
| `4` | Вкладка **Swarm Cluster**: Кластер Docker Swarm (ноды, менеджеры, сервисы, реплики) |
| `Tab` / `→` / `L` | Следующая вкладка |
| `Shift+Tab` / `←` / `H` | Предыдущая вкладка |
| `R` | Немедленное обновление данных |
| `?` | Показать / скрыть подсказку горячих клавиш |
| `Q` / `Esc` / `Ctrl+C` | Выход из приложения |

---

## 🛠️ Флаги командной строки

```text
Использование:
  system-monitor [опции]

Опции:
  -i, --interval <duration>  Интервал автообновления (по умолчанию: 2s, например: -i 1s или -i 5s)
  -t, --tab <1-4>            Начальная вкладка (по умолчанию: 1)
  -v, --version              Показать версию программы
  -h, --help                 Показать справку
```

---

## ⚙️ Команды Makefile

- `make test` — запустить unit-тесты Go (`go test -v ./...`)
- `make build` — скомпилировать бинарный файл (`system-monitor`)
- `make run` — запустить через `go run ./cmd/system-monitor`
- `make version` — показать текущую версию
- `make release-patch` — поднять patch-версию, обновить changelog и создать git-тег
- `make release-minor` — поднять minor-версию, обновить changelog и создать git-тег
- `make release-major` — поднять major-версию, обновить changelog и создать git-тег
- `make clean` — очистить артефакты сборки
