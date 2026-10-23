"""
swarm.py - Модуль мониторинга кластера Docker Swarm (ноды, сервисы, репликация).
"""

import json
import shutil
import subprocess
from typing import Dict, Any, List
from monitors.docker import find_docker_socket, _query_unix_socket


def get_swarm_summary() -> Dict[str, Any]:
    """
    Получение сводной информации о кластере Docker Swarm.
    
    Returns:
        Словарь с состоянием Swarm:
        - is_active: bool
        - status_display: str
        - node_id: str
        - is_manager: bool
        - nodes_count: int
        - managers_count: int
        - nodes: List[Dict]
        - services: List[Dict]
        - text: str
    """
    sock_path = find_docker_socket()
    has_cli = bool(shutil.which("docker"))

    # 1. Попытка через Docker API /info
    if sock_path:
        info_json = _query_unix_socket(sock_path, "/info")
        if info_json:
            try:
                info = json.loads(info_json)
                swarm_info = info.get("Swarm", {})
                state = swarm_info.get("LocalNodeState", "inactive").lower()

                if state != "active":
                    return {
                        "is_active": False,
                        "status_display": "⚪ Не активен",
                        "node_id": "",
                        "is_manager": False,
                        "nodes_count": 0,
                        "managers_count": 0,
                        "nodes": [],
                        "services": [],
                        "text": "Docker Swarm не активирован",
                    }

                node_id = swarm_info.get("NodeID", "")[:12]
                control_avail = bool(swarm_info.get("ControlAvailable", False))
                nodes_count = int(swarm_info.get("Nodes", 1))
                managers_count = int(swarm_info.get("Managers", 1))

                nodes = get_swarm_nodes(sock_path)
                services = get_swarm_services(sock_path)

                return {
                    "is_active": True,
                    "status_display": "🟢 Активен",
                    "node_id": node_id,
                    "is_manager": control_avail,
                    "nodes_count": nodes_count,
                    "managers_count": managers_count,
                    "nodes": nodes,
                    "services": services,
                    "text": f"Кластер Swarm: {nodes_count} нод ({managers_count} менеджеров), {len(services)} сервисов",
                }
            except Exception:
                pass

    # 2. Fallback через CLI
    if has_cli:
        try:
            res = subprocess.run(
                ["docker", "info", "--format", "{{.Swarm.LocalNodeState}} {{.Swarm.ControlAvailable}} {{.Swarm.Nodes}} {{.Swarm.Managers}}"],
                capture_output=True,
                text=True,
                timeout=0.8,
            )
            if res.returncode == 0:
                parts = res.stdout.strip().split()
                if parts and parts[0].lower() == "active":
                    control_avail = parts[1].lower() == "true" if len(parts) > 1 else False
                    nodes_cnt = int(parts[2]) if len(parts) > 2 else 1
                    mgrs_cnt = int(parts[3]) if len(parts) > 3 else 1
                    return {
                        "is_active": True,
                        "status_display": "🟢 Активен",
                        "node_id": "current",
                        "is_manager": control_avail,
                        "nodes_count": nodes_cnt,
                        "managers_count": mgrs_cnt,
                        "nodes": [],
                        "services": [],
                        "text": f"Кластер Swarm: {nodes_cnt} нод ({mgrs_cnt} менеджеров)",
                    }
        except Exception:
            pass

    return {
        "is_active": False,
        "status_display": "⚪ Не активен",
        "node_id": "",
        "is_manager": False,
        "nodes_count": 0,
        "managers_count": 0,
        "nodes": [],
        "services": [],
        "text": "Docker Swarm не активирован",
    }


def get_swarm_nodes(sock_path: str) -> List[Dict[str, Any]]:
    """Получение списка нод кластера."""
    data_text = _query_unix_socket(sock_path, "/nodes")
    nodes = []
    if data_text:
        try:
            raw_nodes = json.loads(data_text)
            for n in raw_nodes:
                spec = n.get("Spec", {})
                desc = n.get("Description", {})
                status = n.get("Status", {})
                manager = n.get("ManagerStatus", {})

                hostname = desc.get("Hostname", "node")
                role = spec.get("Role", "worker")
                state = status.get("State", "unknown")
                is_leader = manager.get("Leader", False)

                nodes.append({
                    "id": n.get("ID", "")[:12],
                    "hostname": hostname,
                    "role": role,
                    "state": state,
                    "is_leader": is_leader,
                    "availability": spec.get("Availability", "active"),
                    "leader_display": "★ Leader" if is_leader else "",
                })
        except Exception:
            pass
    return nodes


def get_swarm_services(sock_path: str) -> List[Dict[str, Any]]:
    """Получение списка сервисов Swarm."""
    data_text = _query_unix_socket(sock_path, "/services")
    services = []
    if data_text:
        try:
            raw_services = json.loads(data_text)
            for s in raw_services:
                spec = s.get("Spec", {})
                name = spec.get("Name", "service")
                task_spec = spec.get("TaskTemplate", {}).get("ContainerSpec", {})
                image = task_spec.get("Image", "").split("@")[0]

                mode_spec = spec.get("Mode", {})
                if "Replicated" in mode_spec:
                    replicas = mode_spec["Replicated"].get("Replicas", 1)
                    mode_str = f"replicated ({replicas})"
                else:
                    mode_str = "global"

                services.append({
                    "id": s.get("ID", "")[:12],
                    "name": name,
                    "image": image,
                    "mode": mode_str,
                })
        except Exception:
            pass
    return services
