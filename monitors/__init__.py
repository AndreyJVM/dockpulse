"""
monitors package
"""
from monitors.system import get_os_info, get_uptime, get_hostname, get_system_summary
from monitors.network import get_default_ip, get_dns_servers, get_network_summary
from monitors.hardware import get_cpu_info, get_ram_info, get_disks_info, get_hardware_summary

__all__ = [
    "get_os_info",
    "get_uptime",
    "get_hostname",
    "get_system_summary",
    "get_default_ip",
    "get_dns_servers",
    "get_network_summary",
    "get_cpu_info",
    "get_ram_info",
    "get_disks_info",
    "get_hardware_summary",
]
