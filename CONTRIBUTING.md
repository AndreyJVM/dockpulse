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
4. Provide a clear description of the problem solved, what changes were made, and how they were tested.

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
4. В описании PR укажите суть изменений, решённую проблему и способ проверки.
