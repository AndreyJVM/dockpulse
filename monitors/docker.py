"""
docker.py - Модуль сбора информации о службе Docker, контейнерах и образах.
"""

import json
import os
import shutil
import socket
import subprocess
from typing import Dict, Any, List, Optional


def find_docker_socket() -> Optional[str]:
    """
    Поиск пути к UNIX-сокету Docker.
    Проверяет переменные окружения и стандартные пути Linux / macOS.
    """
    docker_host = os.environ.get("DOCKER_HOST")
    if docker_host and docker_host.startswith("unix://"):
        path = docker_host[7:]
        if os.path.exists(path):
            return path

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


def _query_unix_socket(sock_path: str, endpoint: str, timeout: float = 0.6) -> Optional[str]:
    """
    Выполнение легковесного HTTP GET-запроса через UNIX сокет.
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
            chunk = s.recv(8192)
            if not chunk:
                break
            response_chunks.append(chunk)

        response_bytes = b"".join(response_chunks)
        response_text = response_bytes.decode("utf-8", errors="replace")

        parts = response_text.split("\r\n\r\n", 1)
        if len(parts) == 2 and ("200 OK" in parts[0] or "HTTP/1.1 200" in parts[0]):
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
    Сводная информация о статусе Docker.
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

    # 1. Попытка через сокет /info
    if sock_path:
        info_json = _query_unix_socket(sock_path, "/info")
        if info_json:
            try:
                info = json.loads(info_json)
                running = int(info.get("ContainersRunning", 0))
                paused = int(info.get("ContainersPaused", 0))
                stopped = int(info.get("ContainersStopped", 0)) + paused
                total = int(info.get("Containers", running + stopped))
                images = int(info.get("Images", 0))
                version = info.get("ServerVersion", "")

                return {
                    "status": "running",
                    "status_display": "🟢 Запущен (Active)",
                    "running_containers": running,
                    "stopped_containers": stopped,
                    "total_containers": total,
                    "images_count": images,
                    "version": version,
                    "text": f"{running} активно, {stopped} остановлено (всего {total})",
                }
            except Exception:
                pass

    # 2. Fallback через CLI
    if has_cli:
        try:
            res = subprocess.run(
                ["docker", "info", "--format", "{{.ContainersRunning}} {{.ContainersStopped}} {{.Containers}} {{.Images}} {{.ServerVersion}}"],
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
                    version = parts[4] if len(parts) > 4 else ""
                    return {
                        "status": "running",
                        "status_display": "🟢 Запущен (Active)",
                        "running_containers": running,
                        "stopped_containers": stopped,
                        "total_containers": total,
                        "images_count": images,
                        "version": version,
                        "text": f"{running} активно, {stopped} остановлено (всего {total})",
                    }
        except Exception:
            pass

    return {
        "status": "stopped",
        "status_display": "⚪ Остановлен (Inactive)",
        "running_containers": 0,
        "stopped_containers": 0,
        "total_containers": 0,
        "images_count": 0,
        "text": "Демон Docker выключен",
    }


def get_docker_containers() -> List[Dict[str, Any]]:
    """
    Получение детального списка всех контейнеров.
    """
    sock_path = find_docker_socket()
    containers = []

    # 1. Попытка через Docker Engine API
    if sock_path:
        data_text = _query_unix_socket(sock_path, "/containers/json?all=1")
        if data_text:
            try:
                raw_list = json.loads(data_text)
                for item in raw_list:
                    names = item.get("Names", ["/unknown"])
                    name = names[0].lstrip("/") if names else "unknown"
                    cid = item.get("Id", "")[:12]
                    image = item.get("Image", "")
                    state = item.get("State", "")
                    status = item.get("Status", "")
                    labels = item.get("Labels", {}) or {}

                    ports_list = []
                    for p in item.get("Ports", []):
                        pub = p.get("PublicPort")
                        priv = p.get("PrivatePort")
                        ptype = p.get("Type", "tcp")
                        if pub:
                            ports_list.append(f"{pub}->{priv}/{ptype}")
                        elif priv:
                            ports_list.append(f"{priv}/{ptype}")
                    ports_str = ", ".join(ports_list) if ports_list else "—"

                    compose_proj = labels.get("com.docker.compose.project", "")
                    compose_svc = labels.get("com.docker.compose.service", "")

                    containers.append({
                        "id": cid,
                        "name": name,
                        "image": image,
                        "state": state,
                        "status": status,
                        "ports": ports_str,
                        "compose_project": compose_proj,
                        "compose_service": compose_svc,
                        "labels": labels,
                    })
                return containers
            except Exception:
                pass

    # 2. Fallback через CLI
    if bool(shutil.which("docker")):
        try:
            res = subprocess.run(
                ["docker", "ps", "-a", "--format", "{{.ID}}\t{{.Names}}\t{{.Image}}\t{{.State}}\t{{.Status}}\t{{.Ports}}\t{{.Labels}}"],
                capture_output=True,
                text=True,
                timeout=1.0,
            )
            if res.returncode == 0 and res.stdout.strip():
                for line in res.stdout.strip().splitlines():
                    parts = line.split("\t")
                    if len(parts) >= 5:
                        cid = parts[0]
                        name = parts[1]
                        image = parts[2]
                        state = parts[3].lower()
                        status = parts[4]
                        ports_str = parts[5] if len(parts) > 5 and parts[5] else "—"
                        labels_str = parts[6] if len(parts) > 6 else ""

                        # Извлечение compose-меток
                        compose_proj = ""
                        compose_svc = ""
                        for pair in labels_str.split(","):
                            if "com.docker.compose.project=" in pair:
                                compose_proj = pair.split("=", 1)[1].strip()
                            elif "com.docker.compose.service=" in pair:
                                compose_svc = pair.split("=", 1)[1].strip()

                        containers.append({
                            "id": cid,
                            "name": name,
                            "image": image,
                            "state": state,
                            "status": status,
                            "ports": ports_str,
                            "compose_project": compose_proj,
                            "compose_service": compose_svc,
                            "labels": {},
                        })
        except Exception:
            pass

    return containers


def get_docker_images() -> List[Dict[str, Any]]:
    """
    Получение списка локальных Docker-образов.
    """
    sock_path = find_docker_socket()
    images = []

    if sock_path:
        data_text = _query_unix_socket(sock_path, "/images/json")
        if data_text:
            try:
                raw_list = json.loads(data_text)
                for item in raw_list:
                    iid = item.get("Id", "").replace("sha256:", "")[:12]
                    tags = item.get("RepoTags") or ["<none>:<none>"]
                    size_mb = item.get("Size", 0) / (1024 * 1024)
                    images.append({
                        "id": iid,
                        "tag": tags[0] if tags else "<none>",
                        "size_mb": size_mb,
                        "size_str": f"{size_mb:.1f} MB",
                    })
                return images
            except Exception:
                pass

    return images
