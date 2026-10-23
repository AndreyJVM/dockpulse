# System & Docker Monitor (Go Edition) ⚡

Сверхбыстрый автономный терминальный TUI-монитор системных ресурсов, **Docker**, **Docker Compose** и кластеров **Docker Swarm**, написанный на **Go (Golang)** с использованием **Bubble Tea** & **Lip Gloss**.

```text
┌─ System & Docker Monitor ───────────────────────────────────────────┐
│ [1: Host Overview]  [2: Docker Containers]  [3: Compose]  [4: Swarm]   │
├─────────────────────────────────────────────────────────────────────┤
│ 🖥️  HOST STATUS       ⚡ PERFORMANCE                                     │
│ Host:    server       CPU (8 cores):   [████████░░░░░░░░] 42.5%         │
│ OS:      Ubuntu 24.04 RAM (6.2/16 GB): [██████████░░░░░░] 38.8%         │
│ Uptime:  4 дн 12:40   NET Traffic:     ↓ 145.2 KB/s   ↑ 24.8 KB/s       │
│                                                                     │
│ 🐳 DOCKER CONTAINERS (Running • 4 active / 2 stopped)               │
│ 📦 DOCKER COMPOSE: 2 активных стека (web-prod, monitoring)          │
│ 🌐 DOCKER SWARM:   Кластер активен (3 ноды, 1 Leader, 5 служб)       │
├─────────────────────────────────────────────────────────────────────┤
│ [1-4] Вкладки  |  [R] Обновить  |  [Q] Выход                        │
└─────────────────────────────────────────────────────────────────────┘
```

---

## ✨ Возможности

- **🖥️ Состояние хоста**: Загрузка CPU с количеством ядер, память RAM, смонтированные дисковые разделы, скорость сети в реальном времени (RX/TX) и Uptime.
- **🐳 Docker Containers**: Прямой опрос unix-сокета (`/var/run/docker.sock`, `~/.docker/run/docker.sock`) без оверхеда SDK. Полный список контейнеров, статусы, образы и порты.
- **📦 Docker Compose**: Автоматическая группировка контейнеров по проектам и сервисам Compose, агрегированный статус стеков (активен, частичный, остановлен).
- **🌐 Docker Swarm**: Мониторинг кластера Swarm — роли нод (Manager/Worker, Leader), реплицированные службы и число реплик.
- **⚡ Интерактивный TUI**: Полноэкранный псевдографический интерфейс Charm Bubble Tea & Lip Gloss с цветными прогресс-барами, навигацией по вкладкам (`1-4`), обновлением по `R` и авто-тиком.
- **🚀 Нулевые зависимости**: Скомпилированный статический бинарный файл без внешних библиотек и рантаймов (<12 МБ). Работает на любом дистрибутиве Linux, macOS и Windows.

---

## 🚀 Быстрый запуск

### 1. Установка готового `.deb` пакета (Debian / Ubuntu / Linux Mint):
```bash
# Для amd64 / x86_64:
wget https://github.com/AndreyJVM/system-monitor/releases/download/v0.3.0/system-monitor_0.3.0_amd64.deb
sudo dpkg -i system-monitor_0.3.0_amd64.deb

# Для arm64 (Raspberry Pi, ARM-серверы):
wget https://github.com/AndreyJVM/system-monitor/releases/download/v0.3.0/system-monitor_0.3.0_arm64.deb
sudo dpkg -i system-monitor_0.3.0_arm64.deb
```

### 2. Запуск готового бинарника (Linux / macOS):
Скачайте архив для вашей ОС со страницы [Releases](https://github.com/AndreyJVM/system-monitor/releases):
```bash
tar -xzf system-monitor_v0.3.0_linux_amd64.tar.gz
./system-monitor
```

### 3. Сборка из исходников (Go):
```bash
git clone https://github.com/AndreyJVM/system-monitor.git
cd system-monitor

# Запуск приложения напрямую
go run ./cmd/system-monitor

# Или сборка исполняемого файла
make build
./system-monitor
```

---

## ⌨️ Управление в TUI:

| Клавиша | Действие |
|---|---|
| `1` | Экран хоста: CPU, RAM, диски, сеть RX/TX, общая сводка Docker |
| `2` | Список контейнеров Docker (ID, Image, State, Ports) |
| `3` | Стеки и проекты Docker Compose |
| `4` | Кластер Docker Swarm (ноды, сервисы, реплики) |
| `R` | Немедленное обновление данных |
| `Q` / `Ctrl+C` | Выход из приложения |

---

## ⚙️ Команды Makefile

- `make test` — запустить unit-тесты Go (`go test -v ./...`)
- `make build` — скомпилировать бинарный файл (`system-monitor`)
- `make run` — запустить через `go run`
- `make version` — показать текущую версию
- `make release-patch` — поднять patch-версию, обновить changelog и создать git-тег
- `make release-minor` — поднять minor-версию, обновить changelog и создать git-тег
- `make release-major` — поднять major-версию, обновить changelog и создать git-тег
- `make clean` — очистить артефакты сборки
