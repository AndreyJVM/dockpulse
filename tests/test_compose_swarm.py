"""
Unit tests for compose and swarm monitors
"""

import unittest
from unittest.mock import patch
from monitors.compose import get_compose_projects
from monitors.swarm import get_swarm_summary, get_swarm_nodes, get_swarm_services


class TestComposeAndSwarm(unittest.TestCase):

    def test_compose_empty_containers(self):
        """Проверка compose при отсутствии контейнеров."""
        with patch("monitors.compose.get_docker_containers", return_value=[]):
            projects = get_compose_projects()
            self.assertEqual(projects, [])

    def test_compose_grouping(self):
        """Проверка группировки контейнеров в compose-проекты."""
        mock_containers = [
            {
                "id": "c1",
                "name": "web_1",
                "image": "nginx",
                "state": "running",
                "status": "Up 2 hours",
                "ports": "80->80/tcp",
                "compose_project": "my_app",
                "compose_service": "web",
                "labels": {},
            },
            {
                "id": "c2",
                "name": "db_1",
                "image": "postgres",
                "state": "running",
                "status": "Up 2 hours",
                "ports": "5432/tcp",
                "compose_project": "my_app",
                "compose_service": "db",
                "labels": {},
            },
        ]
        with patch("monitors.compose.get_docker_containers", return_value=mock_containers):
            projects = get_compose_projects()
            self.assertEqual(len(projects), 1)
            proj = projects[0]
            self.assertEqual(proj["name"], "my_app")
            self.assertEqual(proj["status"], "running")
            self.assertEqual(proj["total_services"], 2)
            self.assertEqual(proj["running_services"], 2)

    def test_swarm_inactive(self):
        """Проверка swarm, когда он не активен."""
        with patch("monitors.swarm.find_docker_socket", return_value=None), \
             patch("shutil.which", return_value=None):
            summary = get_swarm_summary()
            self.assertFalse(summary["is_active"])
            self.assertEqual(summary["nodes"], [])
            self.assertEqual(summary["services"], [])

    def test_swarm_active_mock(self):
        """Проверка swarm при активном кластере."""
        mock_info = json_info = (
            '{"Swarm": {"LocalNodeState": "active", "ControlAvailable": true, '
            '"NodeID": "node123456789", "Nodes": 3, "Managers": 1}}'
        )
        with patch("monitors.swarm.find_docker_socket", return_value="/var/run/docker.sock"), \
             patch("monitors.swarm._query_unix_socket", return_value=mock_info), \
             patch("monitors.swarm.get_swarm_nodes", return_value=[{"id": "n1", "hostname": "mgr1"}]), \
             patch("monitors.swarm.get_swarm_services", return_value=[{"id": "s1", "name": "web"}]):
            summary = get_swarm_summary()
            self.assertTrue(summary["is_active"])
            self.assertTrue(summary["is_manager"])
            self.assertEqual(summary["nodes_count"], 3)
            self.assertEqual(len(summary["nodes"]), 1)
            self.assertEqual(len(summary["services"]), 1)


if __name__ == "__main__":
    unittest.main()
