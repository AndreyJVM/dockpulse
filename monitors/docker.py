"""
docker.py - Модуль сбора информации о статусе Docker и контейнерах.
"""

import json
import os
import shutil
import socket
import subprocess
from typing import Dict, Any, Optional


def find_docker_socket() -> Optional[str]:
    """
    Поиск пути к UNIX-сокету Docker.
    Проверяет переменные окружения и стандартные пути Linux / macOS.
    """
    # 1. Проверка DOCKER_HOST
    docker_host = os.environ.get("DOCKER_HOST")
    if docker_host and docker_host.startswith("unix://"):
        path = docker_host[7:]
        if os.path.exists(path):
            return path

    # 2. Стандартные пути
    candidates = [
        "/var/run/docker.sock",
        "/run/docker.sock",
        os.path.expanduser("~/.docker/run/docker.sock"),
        os.path.expanduser("~/.docker/desktop/docker.sock"),
    ]
    for path in candidates:
        if os.path.exists(path):
            return path

    return None


def _query_unix_socket(sock_path: str, endpoint: str, timeout: float = 0.5) -> Optional[str]:
    """
    Выполнение легковесного HTTP GET-запроса через UNIX сокет без внешних зависимостей.
    """
    s = None
    try:
        s = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        s.settimeout(timeout)
        s.connect(sock_path)
        request = f"GET {endpoint} HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n"
        s.sendall(request.encode("utf-8"))

        response_chunks = []
        while True:
            chunk = s.recv(4096)
            if not chunk:
                break
            response_chunks.append(chunk)

        response_bytes = b"".join(response_chunks)
        response_text = response_bytes.decode("utf-8", errors="replace")

        # Отделяем заголовки от тела ответа
        parts = response_text.split("\r\n\r\n", 1)
        if len(parts) == 2 and "200 OK" in parts[0]:
            return parts[1]
        return None
    except Exception:
        return None
    finally:
        if s is not None:
            try:
                s.close()
            except Exception:
                pass


def get_docker_summary() -> Dict[str, Any]:
    """
    Сбор статуса службы Docker и информации о контейнерах.
    
    Returns:
        Словарь с ключами:
        - status: 'running' | 'stopped' | 'not_installed'
        - status_display: Строка со значком статуса
        - running_containers: int
        - stopped_containers: int
        - total_containers: int
        - images_count: int
        - text: Описание для интерфейса
    """
    has_cli = bool(shutil.which("docker"))
    sock_path = find_docker_socket()

    if not has_cli and not sock_path:
        return {
            "status": "not_installed",
            "status_display": "🔴 Не установлен",
            "running_containers": 0,
            "stopped_containers": 0,
            "total_containers": 0,
            "images_count": 0,
            "text": "Docker не установлен в системе",
        }

    # Если сокет найден, проверяем доступность демона
    if sock_path:
        info_json_text = _query_unix_socket(sock_path, "/info")
        if info_json_text:
            try:
                info = json.loads(info_json_text)
                running = int(info.get("ContainersRunning", 0))
                paused = int(info.get("ContainersPaused", 0))
                stopped = int(info.get("ContainersStopped", 0)) + paused
                total = int(info.get("Containers", running + stopped))
                images = int(info.get("Images", 0))

                return {
                    "status": "running",
                    "status_display": "🟢 Запущен (Active)",
                    "running_containers": running,
                    "stopped_containers": stopped,
                    "total_containers": total,
                    "images_count": images,
                    "text": f"{running} активно, {stopped} остановлено (всего {total})",
                }
            except Exception:
                pass

    # Fallback: быстрая проверка через CLI с таймаутом
    if has_cli:
        try:
            res = subprocess.run(
                ["docker", "info", "--format", "{{.ContainersRunning}} {{.ContainersStopped}} {{.Containers}} {{.Images}}"],
                capture_output=True,
                text=True,
                timeout=0.8,
            )
            if res.returncode == 0:
                parts = res.stdout.strip().split()
                if len(parts) >= 3:
                    running = int(parts[0])
                    stopped = int(parts[1])
                    total = int(parts[2])
                    images = int(parts[3]) if len(parts) > 3 else 0
                    return {
                        "status": "running",
                        "status_display": "🟢 Запущен (Active)",
                        "running_containers": running,
                        "stopped_containers": stopped,
                        "total_containers": total,
                        "images_count": images,
                        "text": f"{running} активно, {stopped} остановлено (всего {total})",
                    }
        except Exception:
            pass

    # Если CLI или сокет есть, но демон не отвечает
    return {
        "status": "stopped",
        "status_display": "⚪ Остановлен (Inactive)",
        "running_containers": 0,
        "stopped_containers": 0,
        "total_containers": 0,
        "images_count": 0,
        "text": "Демон Docker выключен",
    }
