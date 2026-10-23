"""
system.py - Модуль сбора системной информации (ОС, ядро, uptime, хост).
"""

import os
import platform
import re
import socket
import subprocess
import time
from datetime import datetime, timedelta
from typing import Dict, Any


def get_os_info() -> str:
    """
    Получение информации об операционной системе и версии ядра.
    
    Returns:
        Строка вида 'Ubuntu 22.04 LTS\\nKernel: 5.15.0-generic'
    """
    try:
        distro = ""
        # 1. Попытка прочитать /etc/os-release (стандарт для Linux)
        if os.path.exists("/etc/os-release"):
            with open("/etc/os-release", "r", encoding="utf-8", errors="replace") as f:
                for line in f:
                    line = line.strip()
                    if line.startswith("PRETTY_NAME="):
                        distro = line.split("=", 1)[1].strip('"\'')
                        break

        # 2. Попытка через стандартный модуль platform для Python 3.10+
        if not distro and hasattr(platform, "freedesktop_os_release"):
            try:
                os_release = platform.freedesktop_os_release()
                distro = os_release.get("PRETTY_NAME", "")
            except OSError:
                pass

        # 3. Fallback на платформу
        system = platform.system()
        release = platform.release()

        if distro:
            return f"{distro}\nKernel: {release}"
        else:
            return f"{system} {release}\nKernel: {platform.version()[:50]}"
    except Exception as e:
        return f"Error: {str(e)[:50]}"


def get_hostname() -> str:
    """
    Получение сетевого имени компьютера (hostname).
    """
    try:
        return socket.gethostname()
    except Exception as e:
        return f"Error: {str(e)[:50]}"


def get_uptime() -> str:
    """
    Получение времени непрерывной работы системы (Uptime).
    
    Returns:
        Строка вида '2 d 4 h 15 m'
    """
    try:
        uptime_seconds = None

        # 1. Попытка прочитать из /proc/uptime (быстрее и точнее на Linux)
        if os.path.exists("/proc/uptime"):
            try:
                with open("/proc/uptime", "r") as f:
                    uptime_seconds = float(f.readline().split()[0])
            except (IOError, ValueError):
                uptime_seconds = None

        # 2. Попытка через psutil.boot_time()
        if uptime_seconds is None:
            try:
                import psutil
                boot_time = psutil.boot_time()
                uptime_seconds = time.time() - boot_time
            except Exception:
                uptime_seconds = None

        # 3. Fallback для macOS / BSD (для локальной разработки)
        if uptime_seconds is None and platform.system() == "Darwin":
            try:
                out = subprocess.check_output(["sysctl", "-n", "kern.boottime"], text=True)
                match = re.search(r"sec = (\d+)", out)
                if match:
                    uptime_seconds = time.time() - int(match.group(1))
            except Exception:
                pass

        if uptime_seconds is None or uptime_seconds < 0:
            return "N/A"

        delta = timedelta(seconds=int(uptime_seconds))
        days = delta.days
        hours, remainder = divmod(delta.seconds, 3600)
        minutes, _ = divmod(remainder, 60)

        parts = []
        if days > 0:
            parts.append(f"{days} d")
        if hours > 0 or days > 0:
            parts.append(f"{hours} h")
        parts.append(f"{minutes} m")

        return " ".join(parts) if parts else "0 m"
    except Exception as e:
        return f"Error: {str(e)[:50]}"


def get_system_summary() -> Dict[str, Any]:
    """
    Возвращает структурированный словарь с основной информацией о системе.
    """
    return {
        "os_info": get_os_info(),
        "hostname": get_hostname(),
        "uptime": get_uptime(),
        "current_time": datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    }
