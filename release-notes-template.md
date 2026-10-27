# Dockpulse v{VERSION}

### Что нового:
- Переход на строгий профессиональный визуальный стиль без смайликов и эмодзи в CLI, TUI и документации
- Автономный TUI интерфейс на Bubble Tea и Lip Gloss
- Интерактивное управление контейнерами Docker (логи, инспекция, старт/стоп/перезапуск)
- Поддержка мониторинга хоста, Docker, Docker Compose и Docker Swarm
- Кроссплатформенная сборка для Linux (amd64, arm64) и macOS (Apple Silicon, Intel)

### Установка:
```bash
# Debian / Ubuntu / Linux Mint (amd64):
wget https://github.com/{REPO_OWNER}/{REPO_NAME}/releases/download/v{VERSION}/dockpulse_{VERSION}_amd64.deb
sudo dpkg -i dockpulse_{VERSION}_amd64.deb

# Автономный бинарник:
tar -xzf dockpulse_v{VERSION}_{OS}_{ARCH}.tar.gz
./dockpulse
```
