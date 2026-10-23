"""
Unit tests for monitors.docker module
"""

import unittest
from unittest.mock import patch
from monitors.docker import get_docker_summary, find_docker_socket


class TestDockerMonitor(unittest.TestCase):

    def test_get_docker_summary_keys(self):
        """Проверка структуры и типов возвращаемых данных Docker."""
        summary = get_docker_summary()
        self.assertIsInstance(summary, dict)
        self.assertIn("status", summary)
        self.assertIn("status_display", summary)
        self.assertIn("running_containers", summary)
        self.assertIn("stopped_containers", summary)
        self.assertIn("total_containers", summary)
        self.assertIn("images_count", summary)
        self.assertIn("text", summary)

        self.assertIn(summary["status"], ["running", "stopped", "not_installed"])
        self.assertIsInstance(summary["running_containers"], int)
        self.assertIsInstance(summary["stopped_containers"], int)
        self.assertIsInstance(summary["total_containers"], int)
        self.assertIsInstance(summary["images_count"], int)

    @patch("monitors.docker.find_docker_socket", return_value=None)
    @patch("shutil.which", return_value=None)
    def test_docker_not_installed(self, mock_which, mock_socket):
        """Проверка сценария, когда Docker не установлен."""
        summary = get_docker_summary()
        self.assertEqual(summary["status"], "not_installed")
        self.assertIn("Не установлен", summary["status_display"])
        self.assertEqual(summary["running_containers"], 0)

    @patch("monitors.docker.find_docker_socket", return_value="/var/run/docker.sock")
    @patch("monitors.docker._query_unix_socket")
    def test_docker_running_via_socket(self, mock_query, mock_socket):
        """Проверка сценария работающего Docker через сокет."""
        mock_info = (
            '{"Containers": 5, "ContainersRunning": 3, "ContainersPaused": 0, '
            '"ContainersStopped": 2, "Images": 10}'
        )
        mock_query.return_value = mock_info

        summary = get_docker_summary()
        self.assertEqual(summary["status"], "running")
        self.assertEqual(summary["running_containers"], 3)
        self.assertEqual(summary["stopped_containers"], 2)
        self.assertEqual(summary["total_containers"], 5)
        self.assertEqual(summary["images_count"], 10)
        self.assertIn("3 активно", summary["text"])


if __name__ == "__main__":
    unittest.main()
