# System & Docker Monitor (TUI & GUI)

Легковесный монитор системных ресурсов, **Docker**, **Docker Compose** и кластеров **Docker Swarm** с интерактивным терминальным интерфейсом (TUI) и десктопным режимом на GTK 3.

```text
┌─ System & Docker Monitor ───────────────────────────────────────────────┐
│ [1: Host Overview]  [2: Docker Containers]  [3: Compose]  [4: Swarm]   │
├─────────────────────────────────────────────────────────────────────────┤
│ 🖥️  HOST STATUS       ⚡ PERFORMANCE                                     │
│ Host:    laptop       CPU (8 cores):   [████████░░░░░░░░] 42.5%         │
│ OS:      Ubuntu 24.04 RAM (6.2/16 GB): [██████████░░░░░░] 38.8%         │
│ Uptime:  4 d 12 h     NET Traffic:     ↓ 145.2 KB/s   ↑ 24.8 KB/s       │
│                                                                         │
│ 🐳 DOCKER CONTAINERS (Running • 4 active / 2 stopped)                   │
│ 📦 DOCKER COMPOSE: 2 активных стека (web-prod, monitoring)              │
│ 🌐 DOCKER SWARM:   Кластер активен (3 ноды, 1 Leader, 5 сервисов)       │
├─────────────────────────────────────────────────────────────────────────┤
│ [1-4] Вкладки  |  [R] Обновить  |  [Q] Выход                            │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## ✨ Возможности

- **🖥️ Состояние хоста**: Загрузка CPU с количеством ядер, память RAM, смонтированные дисковые разделы, скорость сети в реальном времени (RX/TX KB/s) и Uptime.
- **🐳 Docker Containers**: Полный список локальных контейнеров с их статусами, именами, образами и проброшенными портами.
- **📦 Docker Compose**: Автоматическая группировка контейнеров по проектам и сервисам Compose, статус стеков (активен, частичный, остановлен).
- **🌐 Docker Swarm**: Мониторинг кластера Swarm — роли нод (Manager/Worker, Leader), сервисы и репликация.
- **⚡ Интерактивный TUI**: Быстрая псевдографика прямо в консоли или через SSH с нулевыми задержками и горячими клавишами (`1-4`, `R`, `Q`).
- **🎨 Современный GUI**: Десктопный режим на GTK 3 с HeaderBar и карточным интерфейсом.

---

## 🚀 Быстрый запуск

### Запуск из исходников:
```bash
git clone https://github.com/AndreyJVM/system-monitor.git
cd system-monitor

# Создание и активация виртуального окружения
python3 -m venv venv
source venv/bin/activate

pip install -r requirements.txt

# Интерактивный TUI в терминале (по умолчанию)
python3 main.py

# Десктопный графический режим (GTK 3)
python3 main.py --gui
```

### Установка готового `.deb` пакета (Debian / Ubuntu / Linux Mint):
```bash
wget https://github.com/AndreyJVM/system-monitor/releases/download/v0.1.12/system-monitor_0.1.12_all.deb
sudo dpkg -i system-monitor_0.1.12_all.deb
sudo apt-get install -f

# Запуск из любого терминала
system-monitor
```

---

## ⌨️ Управление в TUI:

| Клавиша | Действие |
|---|---|
| `1` | Экран хоста: CPU, RAM, диски, сеть RX/TX |
| `2` | Список контейнеров Docker (ID, Image, State, Ports) |
| `3` | Стеки и проекты Docker Compose |
| `4` | Кластер Docker Swarm (ноды, сервисы, реплики) |
| `R` | Немедленное обновление данных |
| `Q` / `Ctrl+C` | Выход из приложения |

---

## ⚙️ Команды Makefile

| Команда | Описание |
|---|---|
| `make run` | Запуск приложения в терминале |
| `make test` | Прогон полного набора модульных тестов (`unittest`) |
| `make release-patch` | Поднять patch-версию (`0.1.12` -> `0.1.13`), обновить changelog и тег |
| `make release-minor` | Поднять minor-версию (`0.1.12` -> `0.2.0`), обновить changelog и тег |
| `make build` | Сборка `.deb` пакета через `dpkg-buildpackage` |
| `make clean` | Очистка артефактов сборки |

---

## 📂 Структура проекта

```text
system-monitor/
├── monitors/            # Модули сбора системных и контейнерных метрик
│   ├── system.py        # Информация об ОС, ядре, uptime
│   ├── network.py       # IP-адреса, DNS, скорость сетевого трафика (RX/TX)
│   ├── hardware.py      # CPU, RAM, накопители
│   ├── docker.py        # Служба Docker, контейнеры, образы
│   ├── compose.py       # Проекты и сервисы Docker Compose
│   └── swarm.py         # Кластер Swarm (ноды, сервисы, реплики)
├── tui/                 # Терминальный интерактивный интерфейс (TUI)
│   └── dashboard.py     # Полноэкранный дашборд с вкладками 1-4
├── ui/                  # Графический интерфейс на GTK 3
│   └── main_window.py   # Окно с HeaderBar и карточками
├── tests/               # Модульные тесты
├── main.py              # Точка входа с поддержкой флагов --tui и --gui
├── Makefile             # Автоматизация сборки и релизов
└── VERSION              # Текущая версия приложения
```

---

## 📄 Лицензия

MIT License. Подробности в файле `LICENSE`.
