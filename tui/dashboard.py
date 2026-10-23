"""
dashboard.py - Полнофункциональный терминальный дашборд (TUI)
для мониторинга хоста, Docker, Docker Compose и Docker Swarm.
"""

import os
import select
import shutil
import sys
import termios
import time
import tty
from datetime import datetime
from typing import Dict, Any, List

from monitors.system import get_os_info, get_uptime, get_hostname
from monitors.network import get_default_ip, get_dns_servers, get_network_speed
from monitors.hardware import get_hardware_summary
from monitors.docker import get_docker_summary, get_docker_containers, get_docker_images
from monitors.compose import get_compose_projects
from monitors.swarm import get_swarm_summary


# ANSI Цвета и стили
RESET = "\033[0m"
BOLD = "\033[1m"
DIM = "\033[2m"

CLR_CYAN = "\033[36m"
CLR_BLUE = "\033[34m"
CLR_GREEN = "\033[32m"
CLR_YELLOW = "\033[33m"
CLR_RED = "\033[31m"
CLR_MAGENTA = "\033[35m"
CLR_WHITE = "\033[37m"

BG_HEADER = "\033[44;37m"
BG_CARD = "\033[48;5;236m"
BG_TAB_ACTIVE = "\033[46;30m"
BG_TAB_INACTIVE = "\033[48;5;238;37m"


def format_progress_bar(percent: float, width: int = 24) -> str:
    """Генерация цветного прогресс-бара для терминала."""
    pct = max(0.0, min(100.0, percent))
    filled_len = int(width * pct / 100.0)
    empty_len = width - filled_len

    if pct < 70.0:
        color = CLR_GREEN
    elif pct < 90.0:
        color = CLR_YELLOW
    else:
        color = CLR_RED

    bar = f"{color}{'█' * filled_len}{DIM}{'░' * empty_len}{RESET}"
    return f"[{bar}] {color}{pct:5.1f}%{RESET}"


class TerminalDashboard:
    """
    Интерактивный терминальный интерфейс с вкладками:
    1: Обзор хоста и ресурсов
    2: Контейнеры Docker
    3: Проекты Docker Compose
    4: Кластер Docker Swarm
    """

    def __init__(self):
        self.active_tab = 1
        self.running = True
        self.refresh_interval = 2.0  # секунды
        self.last_refresh = 0.0
        self.cache: Dict[str, Any] = {}

    def fetch_all_data(self):
        """Сбор всех данных из мониторов."""
        self.cache["time"] = datetime.now().strftime("%H:%M:%S")
        self.cache["hostname"] = get_hostname()
        self.cache["os"] = get_os_info().replace("\n", " • ")
        self.cache["uptime"] = get_uptime()
        self.cache["ip"] = get_default_ip()
        self.cache["dns"] = get_dns_servers()
        self.cache["net_speed"] = get_network_speed()
        self.cache["hw"] = get_hardware_summary()
        self.cache["docker"] = get_docker_summary()
        self.cache["containers"] = get_docker_containers()
        self.cache["compose"] = get_compose_projects()
        self.cache["swarm"] = get_swarm_summary()
        self.last_refresh = time.time()

    def render(self):
        """Отрисовка активного экрана."""
        term_cols, term_lines = shutil.get_terminal_size((80, 24))
        buf = []

        # 1. Заголовок и переключатель вкладок
        buf.append(f"{BOLD}{BG_HEADER} System & Docker Monitor {RESET} {DIM}v0.2.0{RESET} "
                   f"{CLR_CYAN}[{self.cache.get('time', '')}]{RESET}")

        tabs = [
            (1, "1: Host Overview"),
            (2, "2: Docker Containers"),
            (3, "3: Compose Stacks"),
            (4, "4: Swarm Cluster"),
        ]

        tabs_line = []
        for tab_id, tab_name in tabs:
            if self.active_tab == tab_id:
                tabs_line.append(f"{BOLD}{BG_TAB_ACTIVE} [{tab_name}] {RESET}")
            else:
                tabs_line.append(f"{BG_TAB_INACTIVE}  {tab_name}  {RESET}")

        buf.append(" ".join(tabs_line))
        buf.append("─" * term_cols)

        # 2. Отрисовка контента в зависимости от активной вкладки
        if self.active_tab == 1:
            buf.extend(self._render_tab_host(term_cols))
        elif self.active_tab == 2:
            buf.extend(self._render_tab_docker(term_cols))
        elif self.active_tab == 3:
            buf.extend(self._render_tab_compose(term_cols))
        elif self.active_tab == 4:
            buf.extend(self._render_tab_swarm(term_cols))

        # 3. Подвал с горячими клавишами
        footer = f"{DIM}Команды: {RESET}{BOLD}[1-4]{RESET} Вкладки  |  " \
                 f"{BOLD}[R]{RESET} Обновить  |  " \
                 f"{BOLD}[Q]{RESET} Выход"

        # Дополняем экран до конца терминала
        rendered_lines = "\n".join(buf)
        total_printed_lines = rendered_lines.count("\n") + 1

        padding = max(0, term_lines - total_printed_lines - 2)
        full_output = rendered_lines + ("\n" * padding) + "\n" + ("─" * term_cols) + "\n" + footer

        # Очистка и позиционирование в начало экрана без мерцания
        sys.stdout.write("\033[H" + full_output)
        sys.stdout.flush()

    def _render_tab_host(self, width: int) -> List[str]:
        """Вкладка 1: Состояние хоста."""
        lines = []
        lines.append(f"{BOLD}{CLR_BLUE}🖥️  СИСТЕМНАЯ ИНФОРМАЦИЯ{RESET}")
        lines.append(f"  {DIM}Хост:{RESET}     {BOLD}{self.cache.get('hostname')}{RESET}   "
                     f"{DIM}Uptime:{RESET} {CLR_CYAN}{self.cache.get('uptime')}{RESET}")
        lines.append(f"  {DIM}ОС/Ядро:{RESET}  {self.cache.get('os')}")
        lines.append("")

        lines.append(f"{BOLD}{CLR_BLUE}⚡  ПРОИЗВОДИТЕЛЬНОСТЬ ХОСТА{RESET}")
        hw = self.cache.get("hw", {})
        cpu = hw.get("cpu", {})
        cpu_pct = float(cpu.get("percent", 0.0))
        lines.append(f"  CPU ({cpu.get('cores', 1)} cores): {format_progress_bar(cpu_pct, 26)}")

        ram = hw.get("ram", {})
        ram_pct = float(ram.get("percent", 0.0))
        used_gb = ram.get("used_gb", 0.0)
        total_gb = ram.get("total_gb", 0.0)
        lines.append(f"  RAM ({used_gb:.1f}/{total_gb:.1f} GB): {format_progress_bar(ram_pct, 26)}")
        lines.append("")

        lines.append(f"{BOLD}{CLR_BLUE}🌐  СЕТЬ И ТРАФИК{RESET}")
        net_speed = self.cache.get("net_speed", {})
        rx = net_speed.get("rx_text", "0 B/s")
        tx = net_speed.get("tx_text", "0 B/s")
        lines.append(f"  IP: {BOLD}{self.cache.get('ip')}{RESET}   DNS: {self.cache.get('dns')}")
        lines.append(f"  Скорость трафика: {CLR_GREEN}↓ RX: {rx}{RESET}   {CLR_BLUE}↑ TX: {tx}{RESET}")
        lines.append("")

        lines.append(f"{BOLD}{CLR_BLUE}💾  ДИСКОВЫЕ РАЗДЕЛЫ{RESET}")
        disks = hw.get("disks", [])
        if disks:
            for d in disks[:4]:
                mp = d.get("mountpoint", "/")
                pct = float(d.get("percent", 0.0))
                du = d.get("used_gb", 0.0)
                dt = d.get("total_gb", 0.0)
                lines.append(f"  {mp:<12} {format_progress_bar(pct, 20)}  ({du:.1f}/{dt:.1f} GB)")
        else:
            lines.append("  Диски недоступны")

        lines.append("")
        docker = self.cache.get("docker", {})
        lines.append(f"{BOLD}{CLR_BLUE}🐳  DOCKER СТАТУС{RESET}: {docker.get('status_display')} • {docker.get('text')}")

        return lines

    def _render_tab_docker(self, width: int) -> List[str]:
        """Вкладка 2: Контейнеры Docker."""
        lines = []
        docker = self.cache.get("docker", {})
        lines.append(f"{BOLD}Служба Docker:{RESET} {docker.get('status_display')}   "
                     f"{DIM}Образы:{RESET} {docker.get('images_count', 0)}")
        lines.append("")

        containers = self.cache.get("containers", [])
        if not containers:
            if docker.get("status") == "running":
                lines.append(f"{DIM}Нет активных или остановленных контейнеров.{RESET}")
            else:
                lines.append(f"{CLR_YELLOW}Демон Docker выключен или недоступен.{RESET}")
            return lines

        lines.append(f"{BOLD}{CLR_WHITE}{'ID':<12} {'NAME':<20} {'IMAGE':<24} {'STATE':<10} {'STATUS':<20}{RESET}")
        lines.append("─" * min(width, 95))

        for c in containers[:16]:
            cid = c.get("id", "")[:10]
            name = c.get("name", "")[:18]
            img = c.get("image", "")[:22]
            state = c.get("state", "").lower()
            status = c.get("status", "")[:18]

            if state == "running":
                state_colored = f"{CLR_GREEN}{state:<10}{RESET}"
            elif state == "paused":
                state_colored = f"{CLR_YELLOW}{state:<10}{RESET}"
            else:
                state_colored = f"{DIM}{state:<10}{RESET}"

            lines.append(f"{cid:<12} {BOLD}{name:<20}{RESET} {img:<24} {state_colored} {status:<20}")

        if len(containers) > 16:
            lines.append(f"{DIM}... и ещё {len(containers) - 16} контейнеров{RESET}")

        return lines

    def _render_tab_compose(self, width: int) -> List[str]:
        """Вкладка 3: Проекты Docker Compose."""
        lines = []
        projects = self.cache.get("compose", [])

        if not projects:
            lines.append(f"{BOLD}{CLR_BLUE}📦  DOCKER COMPOSE ПРОЕКТЫ{RESET}")
            lines.append(f"{DIM}Проекты Docker Compose не обнаружены (нет контейнеров с меткой com.docker.compose.project).{RESET}")
            return lines

        lines.append(f"{BOLD}{CLR_BLUE}📦  АКТИВНЫЕ ПРОЕКТЫ COMPOSE ({len(projects)}){RESET}")
        lines.append("")

        for proj in projects:
            name = proj["name"]
            status_disp = proj["status_display"]
            tot = proj["total_services"]
            run = proj["running_services"]

            lines.append(f"┌─ {BOLD}{name}{RESET}  [{status_disp}]  ({run}/{tot} сервисов поднято)")
            for s in proj["services"]:
                sname = s.get("service")
                sstate = s.get("state")
                simg = s.get("image", "")[:20]
                sports = s.get("ports", "—")

                if sstate == "running":
                    st = f"{CLR_GREEN}● running{RESET}"
                else:
                    st = f"{DIM}○ {sstate}{RESET}"

                lines.append(f"│   • {BOLD}{sname:<15}{RESET} {st:<18} {simg:<22} {sports}")
            lines.append("└──────────────────────────────────────────────────────────")

        return lines

    def _render_tab_swarm(self, width: int) -> List[str]:
        """Вкладка 4: Docker Swarm."""
        lines = []
        swarm = self.cache.get("swarm", {})

        if not swarm.get("is_active"):
            lines.append(f"{BOLD}{CLR_BLUE}🌐  DOCKER SWARM КЛАСТЕР{RESET}")
            lines.append(f"Статус: {swarm.get('status_display')}")
            lines.append("")
            lines.append(f"{DIM}Хост не входит в активный кластер Docker Swarm.{RESET}")
            lines.append(f"Для инициализации выполните: {BOLD}docker swarm init{RESET}")
            return lines

        lines.append(f"{BOLD}{CLR_BLUE}🌐  DOCKER SWARM КЛАСТЕР (Активен){RESET}")
        lines.append(f"  {DIM}Управление:{RESET} {'Менеджер' if swarm.get('is_manager') else 'Воркер'}   "
                     f"{DIM}Всего нод:{RESET} {swarm.get('nodes_count')}   "
                     f"{DIM}Менеджеров:{RESET} {swarm.get('managers_count')}")
        lines.append("")

        nodes = swarm.get("nodes", [])
        if nodes:
            lines.append(f"{BOLD}{'NODE ID':<14} {'HOSTNAME':<20} {'ROLE':<12} {'STATE':<10} {'LEADER':<10}{RESET}")
            lines.append("─" * min(width, 70))
            for n in nodes:
                nid = n.get("id", "")[:10]
                hname = n.get("hostname", "")[:18]
                role = n.get("role", "")
                st = n.get("state", "")
                ldr = n.get("leader_display", "")
                lines.append(f"{nid:<14} {hname:<20} {role:<12} {st:<10} {CLR_YELLOW}{ldr:<10}{RESET}")
        else:
            lines.append(f"{DIM}Информация о нодах кластера доступна на Manager-ноде.{RESET}")

        return lines

    def run(self):
        """Главный цикл терминального приложения."""
        # Переключаем терминал в альтернативный экран и скрываем курсор
        sys.stdout.write("\033[?1049h\033[?25l")
        sys.stdout.flush()

        old_settings = None
        try:
            if sys.stdin.isatty():
                old_settings = termios.tcgetattr(sys.stdin)
                tty.setcbreak(sys.stdin.fileno())

            self.fetch_all_data()

            while self.running:
                now = time.time()
                if now - self.last_refresh >= self.refresh_interval:
                    self.fetch_all_data()

                self.render()

                # Неблокирующее чтение клавиатуры с таймаутом
                rlist, _, _ = select.select([sys.stdin], [], [], 0.3)
                if rlist:
                    ch = sys.stdin.read(1)
                    if ch in ('q', 'Q', '\x03'):  # 'q' или Ctrl+C
                        self.running = False
                    elif ch in ('1', '2', '3', '4'):
                        self.active_tab = int(ch)
                        self.render()
                    elif ch in ('r', 'R'):
                        self.fetch_all_data()
                        self.render()

        finally:
            if old_settings and sys.stdin.isatty():
                termios.tcsetattr(sys.stdin, termios.TCSADRAIN, old_settings)
            # Восстанавливаем основной буфер терминала и курсор
            sys.stdout.write("\033[?1049l\033[?25h")
            sys.stdout.flush()


def run_tui():
    """Точка входа для запуска терминального дашборда."""
    dashboard = TerminalDashboard()
    dashboard.run()
