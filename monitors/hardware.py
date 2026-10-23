"""
hardware.py - Модуль сбора информации об аппаратных ресурсах (CPU, RAM, Swap, Диски).
"""

import shutil
from typing import Dict, Any, List


def get_cpu_info() -> Dict[str, Any]:
    """
    Информация об использовании процессора (CPU).
    """
    try:
        import psutil
        percent = psutil.cpu_percent(interval=0.1)
        count = psutil.cpu_count(logical=True) or 1
        return {
            "percent": percent,
            "cores": count,
            "text": f"{percent:.1f}% ({count} cores)"
        }
    except ImportError:
        return {
            "percent": 0.0,
            "cores": 1,
            "text": "psutil not installed"
        }
    except Exception as e:
        return {
            "percent": 0.0,
            "cores": 1,
            "text": f"Error: {str(e)[:40]}"
        }


def get_ram_info() -> Dict[str, Any]:
    """
    Информация об оперативной памяти (RAM).
    """
    try:
        import psutil
        ram = psutil.virtual_memory()
        percent = ram.percent
        used_gb = ram.used / (1024**3)
        total_gb = ram.total / (1024**3)
        available_gb = ram.available / (1024**3)

        bar_length = 20
        filled = int(bar_length * percent / 100)
        bar = '█' * filled + '░' * (bar_length - filled)

        text = (
            f"[{bar}] {percent:.1f}%\n"
            f"Used: {used_gb:.1f} GB / Total: {total_gb:.1f} GB\n"
            f"Available: {available_gb:.1f} GB"
        )

        return {
            "percent": percent,
            "used_gb": used_gb,
            "total_gb": total_gb,
            "available_gb": available_gb,
            "text": text
        }
    except ImportError:
        return {
            "percent": 0.0,
            "used_gb": 0.0,
            "total_gb": 0.0,
            "available_gb": 0.0,
            "text": "psutil not installed"
        }
    except Exception as e:
        return {
            "percent": 0.0,
            "used_gb": 0.0,
            "total_gb": 0.0,
            "available_gb": 0.0,
            "text": f"Error: {str(e)[:40]}"
        }


def get_disks_info() -> List[Dict[str, Any]]:
    """
    Информация о смонтированных дисках и разделах.
    """
    disks = []
    try:
        import psutil
        partitions = psutil.disk_partitions(all=False)
        for partition in partitions:
            if partition.fstype in ["squashfs", "tmpfs", "devtmpfs", "overlay"]:
                continue
            try:
                usage = psutil.disk_usage(partition.mountpoint)
                percent = usage.percent
                used_gb = usage.used / (1024**3)
                total_gb = usage.total / (1024**3)

                bar_length = 15
                filled = int(bar_length * percent / 100)
                bar = '█' * filled + '░' * (bar_length - filled)

                disks.append({
                    "mountpoint": partition.mountpoint,
                    "device": partition.device,
                    "fstype": partition.fstype,
                    "percent": percent,
                    "used_gb": used_gb,
                    "total_gb": total_gb,
                    "line": f"{partition.mountpoint:15} [{bar}] {percent:5.1f}%  ({used_gb:.1f}/{total_gb:.1f} GB)"
                })
            except (PermissionError, OSError):
                continue
    except ImportError:
        # Fallback с помощью shutil.disk_usage для корневого раздела
        try:
            total, used, free = shutil.disk_usage("/")
            total_gb = total / (1024**3)
            used_gb = used / (1024**3)
            percent = (used / total) * 100 if total > 0 else 0
            disks.append({
                "mountpoint": "/",
                "device": "/",
                "fstype": "default",
                "percent": percent,
                "used_gb": used_gb,
                "total_gb": total_gb,
                "line": f"/               {percent:5.1f}%  ({used_gb:.1f}/{total_gb:.1f} GB)"
            })
        except Exception:
            pass
    except Exception as e:
        disks.append({
            "mountpoint": "/",
            "device": "/",
            "fstype": "error",
            "percent": 0.0,
            "used_gb": 0.0,
            "total_gb": 0.0,
            "line": f"Error: {str(e)[:40]}"
        })

    return disks


def get_hardware_summary() -> Dict[str, Any]:
    """
    Возвращает структурированный словарь с информацией об аппаратных ресурсах.
    """
    disks = get_disks_info()
    if disks:
        disk_lines = [d["line"] for d in disks[:8]]
        if len(disks) > 8:
            disk_lines.append(f"... and {len(disks) - 8} more")
        disks_text = "\n".join(disk_lines)
    else:
        disks_text = "No disk information available"

    return {
        "cpu": get_cpu_info(),
        "ram": get_ram_info(),
        "disks": disks,
        "disks_text": disks_text
    }
