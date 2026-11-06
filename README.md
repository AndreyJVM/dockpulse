<div align="center">

# Dockpulse

[![CI](https://github.com/AndreyJVM/dockpulse/actions/workflows/ci.yml/badge.svg)](https://github.com/AndreyJVM/dockpulse/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/AndreyJVM/dockpulse?color=blue&logo=github)](https://github.com/AndreyJVM/dockpulse/releases)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS-lightgrey)](https://github.com/AndreyJVM/dockpulse/releases)

**A fast, lightweight terminal TUI monitor for Host, Docker, Compose, and Swarm.**<br>
*Быстрый и легковесный терминальный TUI-монитор для хоста, Docker, Compose и Swarm.*

[English](#english) • [Русский](#русский)

---

</div>

## English

### Key Features

- **Host Metrics**: Real-time CPU usage, core counts, RAM, sparkline trend graphs, disk mount points, network traffic (RX/TX), uptime, and top resource-consuming processes.
- **Docker Containers**:
  - Live CPU% and memory stats with dynamic color coding.
  - Interactive sorting by CPU%, RAM, Name, or Status (`s`, `o`).
  - Healthcheck statuses (`[healthy]`, `[unhealthy]`, `[starting]`).
  - Search & filter by name, image, ID, or health status (`/`).
  - Logs viewer with scroll support (`L`).
  - Detailed container inspection modal (`Enter` / `I`).
  - Interactive shell session (`E` — `/bin/bash` or `/bin/sh`) inside container.
  - Power control: Start / Stop (`S`), Restart (`P`), Force Kill (`K` — `SIGKILL`).
  - Safe container removal with confirmation (`x`, `d`).
  - Docker System Prune modal (`C`) with disk reclaim estimates and confirmation.
  - Socket support: Docker daemon, Rootless Docker, Docker Desktop (macOS), and Podman (<2 ms latency).
- **Docker Compose**: Automatic project/service grouping, aggregated stack health, live stack logs (`L`), service control (`U` / `S` / `P`), and stack teardown (`D` / down) with confirmation.
- **Docker Swarm**: Cluster node roles (Leader, Manager, Worker), service listings, replica counts with degradation highlights, and live service logs (`L`).
- **Bilingual Interface (i18n)**: English and Russian language support with automatic locale detection (`LANG`, `LC_*`) and `--lang` flag.
- **Zero Dependencies**: Standalone binary (<12 MB) with zero external runtime dependencies. Runs on Linux and macOS.

### Installation

#### Pre-built `.deb` Package (Debian / Ubuntu / Linux Mint)

```bash
# AMD64 (x86_64)
wget https://github.com/AndreyJVM/dockpulse/releases/latest/download/dockpulse_0.7.0_amd64.deb
sudo dpkg -i dockpulse_0.7.0_amd64.deb

# ARM64
wget https://github.com/AndreyJVM/dockpulse/releases/latest/download/dockpulse_0.7.0_arm64.deb
sudo dpkg -i dockpulse_0.7.0_arm64.deb
```

#### Pre-built Binaries (Linux / macOS)

Download the latest archive from [Releases](https://github.com/AndreyJVM/dockpulse/releases):

```bash
# Linux AMD64
tar -xzf dockpulse_v0.7.0_linux_amd64.tar.gz
./dockpulse

# macOS (Apple Silicon)
tar -xzf dockpulse_v0.7.0_darwin_arm64.tar.gz
./dockpulse
```

#### Build from Source (Go 1.21+)

```bash
git clone https://github.com/AndreyJVM/dockpulse.git
cd dockpulse
make build
./dockpulse
```

### CLI Options

```text
Usage:
  dockpulse [options]

Options:
  -i, --interval <duration>  Auto-refresh interval (default: 2s, e.g. 1s, 500ms)
  -t, --tab <1-4>            Initial active tab (1: Host, 2: Docker, 3: Compose, 4: Swarm) (default: 1)
  -l, --lang <en|ru>         Interface language (default: auto-detect from system locale)
  -v, --version              Show version
  -h, --help                 Show help
```

### Keybindings

| Key | Action |
|---|---|
| `1`..`4` / `Tab` / `←`/`→` | Switch tabs (Host, Docker, Compose, Swarm) |
| `r`, `R` | Force refresh data |
| `↑` / `↓` (or `k` / `j`) | Navigate rows / select items |
| `/` | Search & filter items (`Esc` to reset) |
| `s`, `S`, `o`, `O` | Sort containers (CPU, Memory, Name, Status) |
| `L` | View logs (scroll with `↑`/`↓`/`PgUp`/`PgDn`) |
| `Enter` / `I` | Inspect container details |
| `E` | Attach interactive shell (`/bin/bash` or `/bin/sh`) |
| `S` | Start / Stop container or Compose stack |
| `P` | Restart container or Compose stack |
| `K` | Force Kill container (`SIGKILL`) |
| `x`, `d` | Remove container (with confirmation) |
| `U` | Start Compose stack |
| `D` | Down Compose stack (with confirmation) |
| `C` | Docker System Prune modal |
| `?` | Toggle help modal |
| `q` / `Esc` / `Ctrl+C` | Exit application |

---

## Русский

### Основные возможности

- **Состояние хоста**: Загрузка CPU и ядер, память RAM, динамические спарклайны истории, дисковые разделы, скорость сети в реальном времени (RX/TX), аптайм и топ процессов по потреблению ресурсов.
- **Docker-контейнеры**:
  - Мониторинг CPU% и RAM для каждого контейнера с цветовой индикацией.
  - Интерактивная сортировка по CPU%, RAM, имени или статусу (`s`, `o`).
  - Статусы Healthcheck (`[healthy]`, `[unhealthy]`, `[starting]`).
  - Поиск и фильтрация по имени, образу, ID или статусу здоровья (`/`).
  - Встроенный просмотрщик логов с поддержкой прокрутки (`L`).
  - Детальное окно инспекции контейнера (`Enter` / `I`).
  - Интерактивный терминал (`E` — `/bin/bash` или `/bin/sh`) внутри контейнера.
  - Управление питанием: Старт / Стоп (`S`), Перезапуск (`P`), Принудительное завершение (`K` — `SIGKILL`).
  - Безопасное удаление контейнера с подтверждением (`x`, `d`).
  - Очистка Docker System Prune (`C`) с оценкой освобождаемого места и подтверждением.
  - Поддержка сокетов: стандартный Docker, Rootless Docker, Docker Desktop (macOS) и Podman (<2 мс задержки).
- **Docker Compose**: Автоматическая группировка по проектам и сервисам, статус стеков, просмотр логов стека (`L`), запуск/остановка/рестарт (`U` / `S` / `P`) и сворачивание стека (`D` / down) с подтверждением.
- **Docker Swarm**: Мониторинг нод (Leader, Manager, Worker), список служб, число реплик с индикацией деградации и просмотр логов служб (`L`).
- **Двуязычный интерфейс (i18n)**: Русская и английская локализация с автоматическим определением по окружению системы (`LANG`, `LC_*`) и флагом `--lang`.
- **Нулевые зависимости**: Единый статический бинарный файл (<12 МБ) без внешних рантаймов и библиотек. Работает на Linux и macOS.

### Установка

#### Установка `.deb` пакета (Debian / Ubuntu / Linux Mint)

```bash
# AMD64 (x86_64)
wget https://github.com/AndreyJVM/dockpulse/releases/latest/download/dockpulse_0.7.0_amd64.deb
sudo dpkg -i dockpulse_0.7.0_amd64.deb

# ARM64
wget https://github.com/AndreyJVM/dockpulse/releases/latest/download/dockpulse_0.7.0_arm64.deb
sudo dpkg -i dockpulse_0.7.0_arm64.deb
```

#### Готовые бинарные файлы (Linux / macOS)

Скачайте архив последней версии со страницы [Releases](https://github.com/AndreyJVM/dockpulse/releases):

```bash
# Linux AMD64
tar -xzf dockpulse_v0.7.0_linux_amd64.tar.gz
./dockpulse

# macOS (Apple Silicon)
tar -xzf dockpulse_v0.7.0_darwin_arm64.tar.gz
./dockpulse
```

#### Сборка из исходников (Go 1.21+)

```bash
git clone https://github.com/AndreyJVM/dockpulse.git
cd dockpulse
make build
./dockpulse
```

### Параметры командной строки

```text
Использование:
  dockpulse [опции]

Опции:
  -i, --interval <duration>  Интервал автообновления (по умолчанию: 2s, например: 1s, 500ms)
  -t, --tab <1-4>            Начальная вкладка (1: Host, 2: Docker, 3: Compose, 4: Swarm) (по умолчанию: 1)
  -l, --lang <en|ru>         Язык интерфейса: en или ru (по умолчанию: автоопределение)
  -v, --version              Показать версию программы
  -h, --help                 Показать справку
```

### Горячие клавиши

| Клавиша | Действие |
|---|---|
| `1`..`4` / `Tab` / `←`/`→` | Переключение вкладок (Host, Docker, Compose, Swarm) |
| `r`, `R` | Немедленное обновление данных |
| `↑` / `↓` (или `k` / `j`) | Навигация по строкам / выбор элементов |
| `/` | Поиск и фильтрация (`Esc` для сброса) |
| `s`, `S`, `o`, `O` | Сортировка контейнеров (CPU, Память, Имя, Статус) |
| `L` | Просмотр логов (прокрутка `↑`/`↓`/`PgUp`/`PgDn`) |
| `Enter` / `I` | Инспекция контейнера (детали, тома, сеть, порты, Healthcheck) |
| `E` | Подключение к интерактивному шеллу (`/bin/bash` или `/bin/sh`) |
| `S` | Старт / Остановка контейнера или стека Compose |
| `P` | Перезапуск контейнера или стека Compose |
| `K` | Принудительное завершение (`SIGKILL`) контейнера |
| `x`, `d` | Удаление контейнера (с подтверждением) |
| `U` | Запуск стека Compose |
| `D` | Сворачивание стека Compose (`docker compose down`) с подтверждением |
| `C` | Очистка неиспользуемых ресурсов (Docker System Prune) |
| `?` | Окно справки по горячим клавишам |
| `q` / `Esc` / `Ctrl+C` | Выход из приложения |

---

## Screenshots / Скриншоты

> Place real UI screenshots in [`docs/screenshots/`](docs/screenshots/).  
> Скриншоты интерфейса размещаются в каталоге [`docs/screenshots/`](docs/screenshots/).
