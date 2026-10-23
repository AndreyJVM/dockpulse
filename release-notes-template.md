# System Monitor v{VERSION}

### ✨ Что нового:
- Модульная архитектура (`monitors/` и `ui/`)
- Асинхронный неблокирующий сбор метрик (`threading` + `GLib.idle_add`)
- Мониторинг CPU (загрузка, ядра)
- Отображение времени непрерывной работы (Uptime) и сетевого имени хоста (Hostname)
- Улучшена обработка ошибок сети и DNS
- Поддержка стандартов Debian/Ubuntu пакетирования

### 📦 Установка:
```bash
wget https://github.com/{REPO_OWNER}/{REPO_NAME}/releases/download/v{VERSION}/system-monitor_{VERSION}_all.deb
sudo dpkg -i system-monitor_{VERSION}_all.deb
sudo apt-get install -f
```
