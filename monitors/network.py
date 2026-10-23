"""
network.py - Модуль сбора сетевой информации (IP-адреса, DNS, скорость трафика).
"""

import os
import platform
import socket
import subprocess
import time
from typing import Dict, Any, Tuple, Optional


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
            return ", ".join(dns[:3])
        return "Using system defaults"
    except Exception as e:
        return f"Error: {str(e)[:50]}"


def format_speed(bytes_per_sec: float) -> str:
    """
    Форматирование скорости в человекочитаемый вид (B/s, KB/s, MB/s).
    """
    if bytes_per_sec < 1024:
        return f"{bytes_per_sec:.0f} B/s"
    elif bytes_per_sec < 1024 * 1024:
        return f"{bytes_per_sec / 1024:.1f} KB/s"
    else:
        return f"{bytes_per_sec / (1024 * 1024):.2f} MB/s"


class NetworkSpeedTracker:
    """
    Отслеживание сетевого трафика и расчёт скорости в реальном времени.
    """

    def __init__(self):
        self._last_time: Optional[float] = None
        self._last_sent: Optional[int] = None
        self._last_recv: Optional[int] = None

    def _read_io_counters(self) -> Tuple[Optional[int], Optional[int]]:
        """Получение сырых байт (sent, recv)."""
        # 1. Попытка через psutil
        try:
            import psutil
            counters = psutil.net_io_counters()
            return counters.bytes_sent, counters.bytes_recv
        except Exception:
            pass

        # 2. Linux: чтение /proc/net/dev
        if os.path.exists("/proc/net/dev"):
            try:
                tot_rx = 0
                tot_tx = 0
                with open("/proc/net/dev", "r") as f:
                    for line in f.readlines()[2:]:
                        parts = line.split(":")
                        if len(parts) == 2:
                            iface = parts[0].strip()
                            if iface.startswith(("lo", "docker", "veth")):
                                continue
                            vals = parts[1].split()
                            tot_rx += int(vals[0])
                            tot_tx += int(vals[8])
                return tot_tx, tot_rx
            except Exception:
                pass

        # 3. macOS: чтение netstat
        if platform.system() == "Darwin":
            try:
                out = subprocess.check_output(["netstat", "-ib", "-n"], text=True)
                tot_rx = 0
                tot_tx = 0
                for line in out.splitlines()[1:]:
                    if "<Link" in line and not line.startswith("lo"):
                        parts = line.split()
                        if len(parts) >= 10:
                            tot_rx += int(parts[6])
                            tot_tx += int(parts[9])
                return tot_tx, tot_rx
            except Exception:
                pass

        return None, None

    def get_speed(self) -> Dict[str, Any]:
        """
        Расчёт скорости загрузки (RX) и отдачи (TX).
        """
        now = time.time()
        sent, recv = self._read_io_counters()

        if sent is None or recv is None:
            return {
                "rx_speed": 0.0,
                "tx_speed": 0.0,
                "rx_text": "N/A",
                "tx_text": "N/A",
                "text": "N/A"
            }

        if self._last_time is None or now <= self._last_time:
            self._last_time = now
            self._last_sent = sent
            self._last_recv = recv
            return {
                "rx_speed": 0.0,
                "tx_speed": 0.0,
                "rx_text": "0 B/s",
                "tx_text": "0 B/s",
                "text": "↓ 0 B/s  ↑ 0 B/s"
            }

        dt = now - self._last_time
        rx_diff = max(0, recv - (self._last_recv or recv))
        tx_diff = max(0, sent - (self._last_sent or sent))

        rx_speed = rx_diff / dt if dt > 0 else 0.0
        tx_speed = tx_diff / dt if dt > 0 else 0.0

        self._last_time = now
        self._last_sent = sent
        self._last_recv = recv

        rx_text = format_speed(rx_speed)
        tx_text = format_speed(tx_speed)

        return {
            "rx_speed": rx_speed,
            "tx_speed": tx_speed,
            "rx_text": rx_text,
            "tx_text": tx_text,
            "text": f"↓ {rx_text}  ↑ {tx_text}"
        }


# Глобальный экземпляр трекера
_speed_tracker = NetworkSpeedTracker()


def get_network_speed() -> Dict[str, Any]:
    """
    Получение текущей скорости сетевого трафика.
    """
    return _speed_tracker.get_speed()


def get_network_summary() -> Dict[str, Any]:
    """
    Возвращает структурированный словарь с сетевой информацией.
    """
    return {
        "ip": get_default_ip(),
        "dns": get_dns_servers(),
        "speed": get_network_speed()
    }
