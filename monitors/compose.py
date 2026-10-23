"""
compose.py - Модуль мониторинга проектов и сервисов Docker Compose.
"""

from typing import Dict, Any, List
from monitors.docker import get_docker_containers


def get_compose_projects() -> List[Dict[str, Any]]:
    """
    Группировка Docker-контейнеров по проектам и сервисам Docker Compose.
    
    Returns:
        Список проектов Compose:
        - name: Имя проекта Compose
        - status: 'running' | 'partial' | 'stopped'
        - total_services: int
        - running_services: int
        - services: Список сервисов проекта с их состояниями
    """
    containers = get_docker_containers()
    projects_map: Dict[str, Dict[str, Any]] = {}

    for c in containers:
        proj_name = c.get("compose_project")
        if not proj_name:
            continue

        if proj_name not in projects_map:
            projects_map[proj_name] = {
                "name": proj_name,
                "services": [],
                "working_dir": c.get("labels", {}).get("com.docker.compose.project.working_dir", ""),
                "config_files": c.get("labels", {}).get("com.docker.compose.project.config_files", ""),
            }

        service_name = c.get("compose_service") or c.get("name")
        projects_map[proj_name]["services"].append({
            "service": service_name,
            "container_id": c.get("id"),
            "container_name": c.get("name"),
            "image": c.get("image"),
            "state": c.get("state"),
            "status": c.get("status"),
            "ports": c.get("ports"),
        })

    result = []
    for proj_name, data in sorted(projects_map.items()):
        services = data["services"]
        total = len(services)
        running = sum(1 for s in services if s.get("state") == "running")

        if running == total and total > 0:
            status = "running"
            status_display = "🟢 Запущен"
        elif running > 0:
            status = "partial"
            status_display = "🟡 Частично"
        else:
            status = "stopped"
            status_display = "⚪ Остановлен"

        result.append({
            "name": proj_name,
            "status": status,
            "status_display": status_display,
            "total_services": total,
            "running_services": running,
            "services": services,
            "working_dir": data["working_dir"],
            "config_files": data["config_files"],
        })

    return result
