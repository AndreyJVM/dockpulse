#!/usr/bin/env python3
"""
main.py - Главная точка входа приложения System & Docker Monitor.
Поддерживает терминальный интерактивный режим (TUI) и графический режим (GTK 3).
"""

import sys
import os
import argparse

current_dir = os.path.dirname(os.path.abspath(__file__))
if current_dir not in sys.path:
    sys.path.insert(0, current_dir)


def read_version():
    vfile = os.path.join(current_dir, "VERSION")
    if os.path.exists(vfile):
        with open(vfile, "r", encoding="utf-8") as f:
            return f.read().strip()
    return "0.2.0"


def main():
    parser = argparse.ArgumentParser(
        description="System, Docker, Compose & Swarm Monitor (TUI & GUI)"
    )
    parser.add_argument(
        "--tui", "-t", action="store_true", help="Запуск интерактивного терминального интерфейса (TUI)"
    )
    parser.add_argument(
        "--gui", "-g", action="store_true", help="Запуск десктопного графического интерфейса (GTK 3)"
    )
    parser.add_argument(
        "--version", "-v", action="version", version=f"System Monitor {read_version()}"
    )

    args = parser.parse_args()

    # 1. Принудительный запуск GUI
    if args.gui:
        try:
            from app import SystemMonitorApp
            app = SystemMonitorApp()
            app.run()
            return
        except Exception as e:
            print(f"Ошибка запуска GTK 3 интерфейса: {e}. Переключение на TUI...\n")

    # 2. По умолчанию: в терминале, по SSH или без дисплея запускаем TUI
    if args.tui or sys.stdout.isatty() or ("DISPLAY" not in os.environ and "WAYLAND_DISPLAY" not in os.environ):
        from tui.dashboard import run_tui
        run_tui()
    else:
        # Если запущено из меню десктопа без терминала
        try:
            from app import SystemMonitorApp
            app = SystemMonitorApp()
            app.run()
        except Exception:
            from tui.dashboard import run_tui
            run_tui()


if __name__ == "__main__":
    main()
