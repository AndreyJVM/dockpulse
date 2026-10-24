# System & Docker Monitor (Go Edition)

[![Release](https://img.shields.io/github/v/release/AndreyJVM/system-monitor?color=blue&logo=github)](https://github.com/AndreyJVM/system-monitor/releases)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)](https://github.com/AndreyJVM/system-monitor/releases)

Сверхбыстрый автономный терминальный TUI-монитор системных ресурсов хоста, **Docker**, **Docker Compose** и кластеров **Docker Swarm**, написанный на **Go (Golang)** на базе фреймворков **Bubble Tea** и **Lip Gloss** с возможностью **интерактивного управления контейнерами**.

```text
┌─ System & Docker Monitor ───────────────────────────────────────────────────────────────────────┐
│ [1: Host Overview]  [2: Docker Containers]  [3: Compose]  [4: Swarm]                            │
├─────────────────────────────────────────────────────────────────────────────────────────────────┤
│    DOCKER КОНТЕЙНЕРЫ (ИНТЕРАКТИВНОЕ УПРАВЛЕНИЕ)                                                 │
│   Служба: [Запущен]     Образов: 14    Версия: 27.3.1                                            │
│                                                                                                 │
│    ID           NAME             IMAGE            CPU%     MEM USAGE       STATE    PORTS       │
│   ───────────────────────────────────────────────────────────────────────────────────────────── │
│ ►  e7f829a1b021 web-gateway      nginx:alpine     1.2%     18.4M (3.6%)    running  80->80      │
│    18ac38d990bc db-postgres     postgres:16      0.8%     42.1M (8.2%)    running  5432        │
│    90ef1482ac10 redis-cache     redis:7-alpine   -        -               paused               │
│    45a2789bd012 worker-job      app-worker       -        -               exited               │
├─────────────────────────────────────────────────────────────────────────────────────────────────┤
│ [↑/↓] Выбрать | [/] Поиск | [L] Логи | [Enter/I] Инфо | [S] Старт/Стоп | [P] Рестарт            │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Возможности

- **Состояние хоста в реальном времени**: Загрузка CPU с количеством ядер, память RAM, смонтированные дисковые разделы, скорость сети в реальном времени (RX/TX) и точный аптайм.
- **Топ процессов**: Автоматический вывод наиболее ресурсоемких процессов (PID, имя, CPU %, RSS RAM в МБ).
- **Интерактивное управление Docker-контейнерами**:
  - **Мониторинг ресурсов в реальном времени**: Живой опрос и отображение потребления CPU (%) и оперативной памяти (МБ / %) для каждого работающего контейнера с цветовой индикацией прямо в таблице и в окне инспекции.
  - Навигация по списку контейнеров клавишами `↑` / `↓` или `j` / `k`.
  - **Просмотр логов (`L`)**: встроенный модальный просмотрщик последних логов с прокруткой (`↑`/`↓`/`PgUp`/`PgDn`).
  - **Инспекция (`Enter` / `I`)**: всплывающее окно со всеми деталями (IP-адрес, шлюз, MAC, команда запуска, примонтированные тома).
  - **Управление питанием**: быстрый старт / остановка (`S`) и перезапуск (`P`) с мгновенным обновлением статуса.
  - Поддержка стандартного Docker, Rootless Docker, Docker Desktop (macOS) и Podman через unix-сокеты (<2 мс задержки).
- **Docker Compose**: Автоматическая группировка контейнеров по проектам и сервисам Compose, агрегированный статус стеков (активен, частичный, остановлен).
- **Docker Swarm**: Мониторинг кластера Swarm — роли нод (Manager/Worker, Leader), реплицированные службы и число реплик.
- **Интерактивный TUI**: Полноэкранный псевдографический интерфейс на Charm Bubble Tea & Lip Gloss с цветными прогресс-барами, навигацией по вкладкам (`1-4`, `Tab`, стрелки), всплывающей справкой (`?`) и авто-тиком.
- **Нулевые зависимости**: Скомпилированный статический бинарный файл без внешних библиотек и рантаймов (<12 МБ). Работает на любом дистрибутиве Linux, macOS (Apple Silicon и Intel) и Windows.

---

## Установка и быстрый запуск

### Вариант 1. Установка готового `.deb` пакета (Debian / Ubuntu / Linux Mint):
```bash
# Для x86_64 / amd64:
wget https://github.com/AndreyJVM/system-monitor/releases/download/v0.4.1/system-monitor_0.4.1_amd64.deb
sudo dpkg -i system-monitor_0.4.1_amd64.deb

# Для ARM64 (Raspberry Pi, ARM-серверы):
wget https://github.com/AndreyJVM/system-monitor/releases/download/v0.4.1/system-monitor_0.4.1_arm64.deb
sudo dpkg -i system-monitor_0.4.1_arm64.deb

# Запуск из любого терминала
system-monitor
```

### Вариант 2. Запуск готового бинарника (Linux / macOS):
Скачайте архив со страницы [Releases](https://github.com/AndreyJVM/system-monitor/releases):
```bash
# Linux amd64
tar -xzf system-monitor_v0.4.1_linux_amd64.tar.gz
./system-monitor

# macOS (Apple Silicon)
tar -xzf system-monitor_v0.4.1_darwin_arm64.tar.gz
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

## Управление в TUI

### Навигация по вкладкам:
| Клавиша | Действие |
|---|---|
| `1` | Вкладка **Host Overview**: CPU, RAM, диски, сеть RX/TX, топ процессов, сводка Docker |
| `2` | Вкладка **Docker Containers**: Интерактивный список контейнеров и управление |
| `3` | Вкладка **Compose Stacks**: Стеки и сервисы Docker Compose |
| `4` | Вкладка **Swarm Cluster**: Кластер Docker Swarm (ноды, менеджеры, сервисы, реплики) |
| `Tab` / `→` | Следующая вкладка |
| `Shift+Tab` / `←` | Предыдущая вкладка |
| `R` | Немедленное обновление данных |
| `?` | Показать / скрыть окно подсказок |
| `Q` / `Esc` / `Ctrl+C` | Выход из приложения |

### Интерактивное управление контейнерами (на вкладке 2):
| Клавиша | Действие |
|---|---|
| `↑` / `↓` (или `k` / `j`) | Выбор контейнера курсором `►` |
| `/` | **Поиск и фильтрация**: фильтрация списка по имени, образу или ID (`Esc` — сброс) |
| `L` | **Просмотр логов** выбранного контейнера (прокрутка `↑`/`↓`/`PgUp`/`PgDn`) |
| `Enter` / `I` | **Инспекция контейнера** (детальное окно: IP, порты, тома, сеть, команда) |
| `S` | **Старт / Стоп**: запустить остановленный или остановить работающий контейнер |
| `P` | **Перезапуск (Restart)** контейнера |
| `Esc` / `Q` | Закрыть модальное окно логов или инспекции |

---

## Флаги командной строки

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

## Команды Makefile

- `make test` — запустить unit-тесты Go (`go test -v ./...`)
- `make build` — скомпилировать бинарный файл (`system-monitor`)
- `make run` — запустить через `go run ./cmd/system-monitor`
- `make version` — показать текущую версию
- `make release-patch` — поднять patch-версию, обновить changelog и создать git-тег
- `make release-minor` — поднять minor-версию, обновить changelog и создать git-тег
- `make release-major` — поднять major-версию, обновить changelog и создать git-тег
- `make clean` — очистить артефакты сборки
