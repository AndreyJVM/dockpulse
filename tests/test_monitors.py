"""
Tests for monitors package
"""
import unittest
from monitors.system import get_os_info, get_hostname, get_uptime, get_system_summary
from monitors.network import (
    get_default_ip,
    get_dns_servers,
    get_network_speed,
    get_network_summary,
    format_speed,
)
from monitors.hardware import get_cpu_info, get_ram_info, get_disks_info, get_hardware_summary


class TestMonitors(unittest.TestCase):

    def test_system_info(self):
        os_info = get_os_info()
        self.assertIsInstance(os_info, str)
        self.assertTrue(len(os_info) > 0)

        hostname = get_hostname()
        self.assertIsInstance(hostname, str)
        self.assertTrue(len(hostname) > 0)

        uptime = get_uptime()
        self.assertIsInstance(uptime, str)
        self.assertTrue(len(uptime) > 0)

        summary = get_system_summary()
        self.assertIn("os_info", summary)
        self.assertIn("hostname", summary)
        self.assertIn("uptime", summary)
        self.assertIn("current_time", summary)

    def test_network_info(self):
        ip = get_default_ip()
        self.assertIsInstance(ip, str)
        self.assertTrue(len(ip) > 0)

        dns = get_dns_servers()
        self.assertIsInstance(dns, str)
        self.assertTrue(len(dns) > 0)

        summary = get_network_summary()
        self.assertIn("ip", summary)
        self.assertIn("dns", summary)
        self.assertIn("speed", summary)

        speed = get_network_speed()
        self.assertIsInstance(speed, dict)
        self.assertIn("rx_speed", speed)
        self.assertIn("tx_speed", speed)
        self.assertIn("rx_text", speed)
        self.assertIn("tx_text", speed)
        self.assertIn("text", speed)

    def test_format_speed(self):
        self.assertEqual(format_speed(500), "500 B/s")
        self.assertEqual(format_speed(1024), "1.0 KB/s")
        self.assertEqual(format_speed(1024 * 1024), "1.00 MB/s")

    def test_hardware_info(self):
        cpu = get_cpu_info()
        self.assertIsInstance(cpu, dict)
        self.assertIn("percent", cpu)
        self.assertIn("cores", cpu)
        self.assertIn("text", cpu)

        ram = get_ram_info()
        self.assertIsInstance(ram, dict)
        self.assertIn("percent", ram)
        self.assertIn("text", ram)

        disks = get_disks_info()
        self.assertIsInstance(disks, list)

        summary = get_hardware_summary()
        self.assertIn("cpu", summary)
        self.assertIn("ram", summary)
        self.assertIn("disks", summary)
        self.assertIn("disks_text", summary)


if __name__ == "__main__":
    unittest.main()
