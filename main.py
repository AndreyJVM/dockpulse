#!/usr/bin/env python3
"""
main.py - Главная точка входа приложения System Monitor.
"""

import sys
import os

# Гарантируем корректный импорт модулей при локальном запуске
current_dir = os.path.dirname(os.path.abspath(__file__))
if current_dir not in sys.path:
    sys.path.insert(0, current_dir)

from app import SystemMonitorApp


def main():
    app = SystemMonitorApp()
    app.run()


if __name__ == "__main__":
    main()
