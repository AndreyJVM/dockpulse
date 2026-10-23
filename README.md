# System Monitor

Легковесное графическое приложение на Python и GTK 3 для мониторинга системных ресурсов и информации о компьютере в Linux.

![System Monitor](images/main-window.png)

## ✨ Возможности

- **Информация об ОС**: Дистрибутив, версия ядра Linux, имя хоста и время непрерывной работы (Uptime).
- **Сетевая информация**: Текущий IP-адрес, DNS-серверы и скорость передачи трафика в реальном времени (RX / TX в KB/s или MB/s).
- **Наглядные индикаторы (ProgressBar)**: Нативные цветные индикаторы загрузки CPU, оперативной памяти и дисков с динамической цветовой индикацией (зеленый / оранжевый / красный).
- **Мониторинг накопителей**: Список смонтированных разделов диска, используемый и общий объём, процент заполнения.
- **Использование памяти**: Мониторинг RAM (использовано, доступно, общий объем).
- **Мониторинг процессора (CPU)**: Общая загрузка и количество ядер.
- **Автоматическое и ручное обновление**: Фоновое неблокирующее обновление данных и кнопка мгновенного обновления.

---

## 📋 Системные требования

- **ОС**: Linux (Ubuntu, Debian, Linux Mint или производные)
- **Python**: 3.8 или новее
- **GTK**: 3.0+
- **Системные зависимости** (Debian/Ubuntu):
  ```bash
  sudo apt-get update
  sudo apt-get install python3 python3-gi python3-gi-cairo gir1.2-gtk-3.0 python3-psutil
  ```

---

## 🚀 Установка

### Установка готового `.deb` пакета (Debian / Ubuntu / Mint)

Загрузите `.deb` пакет со страницы [Releases](https://github.com/AndreyJVM/system-monitor/releases) или выполните:

```bash
# Загрузка последней версии пакета
wget https://github.com/AndreyJVM/system-monitor/releases/download/v0.1.9/system-monitor_0.1.9_all.deb

# Установка пакета
sudo dpkg -i system-monitor_0.1.9_all.deb

# Установка недостающих зависимостей (если требуется)
sudo apt-get install -f
```

После установки приложение доступно в системном меню или по команде `system-monitor`.

---

## 🛠️ Запуск из исходников (Разработка)

```bash
# Клонирование репозитория
git clone https://github.com/AndreyJVM/system-monitor.git
cd system-monitor

# Создание и активация виртуального окружения
python3 -m venv venv
source venv/bin/activate

# Установка зависимостей Python
pip install -r requirements.txt

# Запуск приложения
python3 main.py
```

---

## ⚙️ Управление проектом (Makefile)

В проекте настроен `Makefile` для автоматизации сборки и развертывания:

| Команда | Описание |
|---|---|
| `make run` | Запустить приложение локально |
| `make test` | Запустить модульные тесты (`unittest`) |
| `make release-patch` | Увеличить patch-версию, обновить changelog и создать релиз-тег |
| `make release-minor` | Увеличить minor-версию, обновить changelog и создать релиз-тег |
| `make build` | Собрать `.deb` пакет через `dpkg-buildpackage` (на Linux) |
| `make install` | Собрать и установить пакет в систему (Debian/Ubuntu) |
| `make clean` | Очистить временные файлы сборки |
| `make version` | Показать текущую версию проекта |

---

## 📂 Структура проекта

```text
system-monitor/
├── monitors/            # Модули сбора системных метрик
│   ├── system.py        # Информация об ОС, ядре, uptime
│   ├── network.py       # IP-адреса, DNS, скорость сетевого трафика (RX/TX)
│   └── hardware.py      # CPU, RAM, накопители
├── ui/                  # Графический интерфейс
│   └── main_window.py   # Главное окно на GTK 3 с ProgressBar
├── tests/               # Модульные тесты
├── images/              # Скриншоты и графические материалы
├── icon.png             # Иконка приложения
├── main.py              # Точка входа в приложение
├── requirements.txt     # Зависимости Python
├── system-monitor.desktop # Десктоп-ярлык приложения
├── Makefile             # Команды автоматизации сборки и релиза
└── VERSION              # Текущая версия приложения
```

---

## 📄 Лицензия

Распространяется под лицензией MIT. Подробности в файле `LICENSE`.
