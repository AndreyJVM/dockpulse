"""
monitors package
"""
from monitors.system import get_os_info, get_uptime, get_hostname, get_system_summary
from monitors.network import (
    get_default_ip,
    get_dns_servers,
    get_network_speed,
    get_network_summary,
    format_speed,
)
from monitors.hardware import get_cpu_info, get_ram_info, get_disks_info, get_hardware_summary
from monitors.docker import get_docker_summary, get_docker_containers, get_docker_images, find_docker_socket
from monitors.compose import get_compose_projects
from monitors.swarm import get_swarm_summary, get_swarm_nodes, get_swarm_services

__all__ = [
    "get_os_info",
    "get_uptime",
    "get_hostname",
    "get_system_summary",
    "get_default_ip",
    "get_dns_servers",
    "get_network_speed",
    "get_network_summary",
    "format_speed",
    "get_cpu_info",
    "get_ram_info",
    "get_disks_info",
    "get_hardware_summary",
    "get_docker_summary",
    "get_docker_containers",
    "get_docker_images",
    "find_docker_socket",
    "get_compose_projects",
    "get_swarm_summary",
    "get_swarm_nodes",
    "get_swarm_services",
]
