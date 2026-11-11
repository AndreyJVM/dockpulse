# Contributing to Dockpulse / Руководство по участию в разработке

[English](#english) • [Русский](#русский)

---

## English

Thank you for your interest in contributing to **Dockpulse**! We welcome all contributions: bug reports, documentation enhancements, feature proposals, and pull requests.

### Prerequisites

- **Go**: 1.27.3 or higher
- **Git**
- **Docker** (optional, recommended for testing Docker/Compose/Swarm integrations)
- **revive** linter (optional, install with `go install github.com/mgechev/revive@latest`)

### Getting Started

1. **Fork** the repository on GitHub:
   [https://github.com/AndreyJVM/dockpulse](https://github.com/AndreyJVM/dockpulse)

2. **Clone** your fork locally and configure the `upstream` remote:
   ```bash
   git clone git@github.com:<your-username>/dockpulse.git
   cd dockpulse
   git remote add upstream https://github.com/AndreyJVM/dockpulse.git
   git fetch upstream
   ```

3. **Create a branch** for your changes starting from `upstream/main`:
   ```bash
   git checkout -b feat/my-new-feature
   # or for bug fixes:
   git checkout -b fix/issue-description
   ```

### Project Architecture & Codebase Structure

Dockpulse is structured into clear and modular packages:

```text
dockpulse/
├── cmd/
│   └── dockpulse/          # Application entrypoint (main.go), CLI flags, and initialization
├── internal/               # Core internal packages (not importable by external projects)
│   ├── monitor/            # Monitoring engine and Docker engine integrations
│   │   ├── docker*.go      # Docker client, container stats, logs, inspect, exec, prune
│   │   ├── compose*.go     # Docker Compose service discovery and lifecycle
│   │   ├── swarm*.go       # Docker Swarm cluster nodes and services monitoring
│   │   └── system*.go      # Host system metrics (CPU, RAM, Disk, Network)
│   ├── tui/                # Terminal User Interface (Bubble Tea & Lip Gloss)
│   │   ├── views/          # Primary views (Docker containers, Compose, Swarm, Host)
│   │   ├── modals/         # Interactive dialogs (logs, inspect, exec, prune, help)
│   │   ├── components/     # Reusable UI widgets (progress bar, sparkline, tables)
│   │   ├── theme/          # Color schemes, styles, and borders
│   │   └── update*.go      # Event handling, keybindings, and command dispatchers
│   └── i18n/               # Internationalization engine and dictionaries (EN / RU)
├── docs/                   # Documentation and screenshots
└── debian/                 # Debian/Ubuntu packaging metadata
```

**Where to make your changes:**
- Adding or modifying metrics/Docker logic? Work in `internal/monitor/`.
- Adding a new dashboard view? Work in `internal/tui/views/`.
- Adding a popup modal or action? Work in `internal/tui/modals/`.
- Adding UI components or widgets? Work in `internal/tui/components/`.
- Updating UI strings or translations? Update `internal/i18n/dictionary.go`.

### Development & Testing

Use the provided `Makefile` targets to build, test, and verify your changes:

```bash
# Build standalone binary
make build

# Run unit tests
make test

# Run code style and quality checks (go vet + revive)
make lint

# Run locally in development mode
make run
```

### Commit Message Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification (see also `commit-template.txt`):

- `feat(...)`: New feature or capability
- `fix(...)`: Bug fix
- `docs(...)`: Documentation changes
- `refactor(...)`: Code refactoring without behavioral changes
- `test(...)`: Adding or updating tests
- `chore(...)`: Maintenance, dependency updates, version bumps

Example:
```text
feat(tui): add container memory limit indicator
docs: add contributing guide
fix(monitor): handle missing docker socket gracefully
```

### Submitting a Pull Request

1. Make sure all tests and linter checks pass (`make test && make lint`).
2. Push your branch to your fork:
   ```bash
   git push origin feat/my-new-feature
   ```
3. Open a Pull Request against the `main` branch of `AndreyJVM/dockpulse`.
4. Fill out the PR template with a clear description of the problem solved, what changes were made, and how they were tested.

### Automated Dependency Updates (Dependabot)

The repository uses **Dependabot** to keep dependencies and GitHub Actions workflows secure and up to date:
- Go module updates are grouped weekly.
- GitHub Actions updates are scheduled monthly.
- Automated CI workflows verify every dependency bump.

---

## Русский

Спасибо за интерес к разработке **Dockpulse**! Мы рады любым предложениям: сообщениям об ошибках, улучшениям документации, новым функциям и Pull Request'ам.

### Требования к окружению

- **Go**: версии 1.27.3 или выше
- **Git**
- **Docker** (опционально, рекомендуется для проверки интеграций с Docker, Compose и Swarm)
- Линтер **revive** (опционально, установка: `go install github.com/mgechev/revive@latest`)

### Начало работы

1. **Сделайте Fork** репозитория на GitHub:
   [https://github.com/AndreyJVM/dockpulse](https://github.com/AndreyJVM/dockpulse)

2. **Клонируйте** свой форк локально и настройте удаленный репозиторий `upstream`:
   ```bash
   git clone git@github.com:<your-username>/dockpulse.git
   cd dockpulse
   git remote add upstream https://github.com/AndreyJVM/dockpulse.git
   git fetch upstream
   ```

3. **Создайте ветку** для ваших изменений от `upstream/main`:
   ```bash
   git checkout -b feat/my-new-feature
   # или для исправлений:
   git checkout -b fix/issue-description
   ```

### Архитектура проекта и структура каталогов

Проект разбит на логически изолированные модули:

```text
dockpulse/
├── cmd/
│   └── dockpulse/          # Точка входа в программу (main.go), парсинг CLI-флагов
├── internal/               # Внутренние пакеты приложения
│   ├── monitor/            # Движок мониторинга и интеграция с Docker
│   │   ├── docker*.go      # Клиент Docker API, статистика, логи, инспекция, exec, prune
│   │   ├── compose*.go     # Обнаружение проектов Compose и управление сервисами
│   │   ├── swarm*.go       # Мониторинг нод и сервисов Docker Swarm
│   │   └── system*.go      # Метрики хоста (CPU, RAM, диск, сеть)
│   ├── tui/                # Терминальный интерфейс (Bubble Tea и Lip Gloss)
│   │   ├── views/          # Основные экраны (Docker, Compose, Swarm, Хост)
│   │   ├── modals/         # Диалоговые окна (логи, инспекция, exec, подтверждения)
│   │   ├── components/     # Переиспользуемые элементы UI (графики, прогресс-бары)
│   │   ├── theme/          # Темы оформления, палитра цветов и рамки
│   │   └── update*.go      # Обработка событий ввода и диспетчеризация команд
│   └── i18n/               # Подсистема локализации и словари (RU / EN)
├── docs/                   # Документация, скриншоты и руководства
└── debian/                 # Файлы сборки пакетов Debian/Ubuntu
```

**Куда вносить изменения:**
- Новые метрики или методы работы с Docker? Смотрите `internal/monitor/`.
- Новый экран интерфейса? Смотрите `internal/tui/views/`.
- Новое всплывающее окно или диалог? Смотрите `internal/tui/modals/`.
- Новый виджет или элемент интерфейса? Смотрите `internal/tui/components/`.
- Перевод или локализованные фразы? Обновляйте `internal/i18n/dictionary.go`.

### Разработка и тестирование

Для сборки, тестирования и проверки кода используйте команды `Makefile`:

```bash
# Сборка автономного бинарника
make build

# Запуск модульных тестов
make test

# Проверка стиля и качества кода (go vet + revive)
make lint

# Локальный запуск приложения
make run
```

### Формат сообщений коммитов

Мы придерживаемся соглашения [Conventional Commits](https://www.conventionalcommits.org/) (см. также `commit-template.txt`):

- `feat(...)`: Добавление новой функциональности
- `fix(...)`: Исправление ошибки
- `docs(...)`: Изменения в документации
- `refactor(...)`: Рефакторинг кода без изменения поведения
- `test(...)`: Добавление или обновление тестов
- `chore(...)`: Обслуживание, обновление зависимостей, версионирование

Примеры:
```text
feat(tui): add container memory limit indicator
docs: add contributing guide
fix(monitor): handle missing docker socket gracefully
```

### Отправка Pull Request

1. Убедитесь, что тесты и линтер проходят успешно (`make test && make lint`).
2. Отправьте ветку в свой форк:
   ```bash
   git push origin feat/my-new-feature
   ```
3. Откройте Pull Request в ветку `main` репозитория `AndreyJVM/dockpulse`.
4. Заполните шаблон PR: укажите суть изменений, решённую проблему и способ проверки.

### Автоматическое обновление зависимостей (Dependabot)

В репозитории настроен **Dependabot**, который помогает поддерживать библиотеки и GitHub Actions в актуальном и безопасном состоянии:
- Зависимости Go обновляются с еженедельной группировкой в один общий PR.
- Версии GitHub Actions проверяются ежемесячно.
- Каждый PR от Dependabot автоматически тестируется через CI.
