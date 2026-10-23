"""
network.py - Модуль сбора сетевой информации (IP-адреса, DNS, интерфейсы).
"""

import os
import platform
import socket
import subprocess
from typing import Dict, Any, List


def get_default_ip() -> str:
    """
    Получение основного исходящего IP-адреса с защитой от зависания по таймауту.
    """
    # 1. Быстрая попытка через UDP сокет с таймаутом 1 секунда
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.settimeout(1.0)
        s.connect(("8.8.8.8", 80))
        ip = s.getsockname()[0]
        s.close()
        return ip
    except Exception:
        pass

    # 2. Попытка через перебор интерфейсов psutil
    try:
        import psutil
        interfaces = psutil.net_if_addrs()
        for iface_name, addrs in interfaces.items():
            if iface_name.startswith(("lo", "docker", "veth", "virbr")):
                continue
            for addr in addrs:
                if addr.family == socket.AF_INET and not addr.address.startswith("127."):
                    return f"{addr.address} ({iface_name})"
    except Exception:
        pass

    # 3. Fallback через имя хоста
    try:
        hostname = socket.gethostname()
        ip = socket.gethostbyname(hostname)
        if ip and not ip.startswith("127."):
            return f"{ip} (local)"
        return ip
    except Exception as e:
        return f"Error: {str(e)[:50]}"


def get_dns_servers() -> str:
    """
    Получение списка активных DNS-серверов.
    """
    dns = []
    try:
        # 1. Попытка чтения /etc/resolv.conf (Linux / Unix)
        if os.path.exists("/etc/resolv.conf"):
            with open("/etc/resolv.conf", "r", encoding="utf-8", errors="replace") as f:
                for line in f:
                    line = line.strip()
                    if line.startswith("nameserver") and not line.startswith("#"):
                        parts = line.split()
                        if len(parts) > 1 and parts[1] not in dns:
                            dns.append(parts[1])

        # 2. Проверка systemd-resolved, если в resolv.conf только 127.0.0.53
        if (not dns or dns == ["127.0.0.53"]) and os.path.exists("/run/systemd/resolve/resolv.conf"):
            with open("/run/systemd/resolve/resolv.conf", "r", encoding="utf-8", errors="replace") as f:
                for line in f:
                    line = line.strip()
                    if line.startswith("nameserver"):
                        parts = line.split()
                        if len(parts) > 1 and parts[1] not in dns:
                            dns.append(parts[1])

        # 3. Fallback для macOS
        if not dns and platform.system() == "Darwin":
            try:
                out = subprocess.check_output(["scutil", "--dns"], text=True)
                for line in out.splitlines():
                    line = line.strip()
                    if line.startswith("nameserver[0]"):
                        parts = line.split(":")
                        if len(parts) > 1:
                            val = parts[1].strip()
                            if val not in dns:
                                dns.append(val)
            except Exception:
                pass

        if dns:
            # Показываем не более 3 серверов
            return ", ".join(dns[:3])
        return "Using system defaults"
    except Exception as e:
        return f"Error: {str(e)[:50]}"


def get_network_summary() -> Dict[str, Any]:
    """
    Возвращает структурированный словарь с сетевой информацией.
    """
    return {
        "ip": get_default_ip(),
        "dns": get_dns_servers()
    }
