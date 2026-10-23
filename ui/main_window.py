"""
main_window.py - Современный графический интерфейс на GTK 3 с карточным дизайном, HeaderBar и мониторингом Docker.
"""

import os
import threading
from datetime import datetime
from typing import Tuple

import gi
gi.require_version('Gtk', '3.0')
from gi.repository import Gtk, GLib, Gdk

from monitors.system import get_os_info, get_uptime, get_hostname
from monitors.network import get_default_ip, get_dns_servers, get_network_speed
from monitors.hardware import get_hardware_summary
from monitors.docker import get_docker_summary


def get_version() -> str:
    """Чтение версии приложения из файла VERSION."""
    try:
        cur_dir = os.path.dirname(os.path.abspath(__file__))
        version_file = os.path.join(os.path.dirname(cur_dir), "VERSION")
        if os.path.exists(version_file):
            with open(version_file, "r", encoding="utf-8") as f:
                return f.read().strip()
    except Exception:
        pass
    return "0.1.11"


CSS_DATA = """
window {
    background-color: @theme_bg_color;
}

headerbar {
    padding: 4px 8px;
}

.card {
    background-color: rgba(128, 128, 128, 0.08);
    border: 1px solid rgba(128, 128, 128, 0.18);
    border-radius: 12px;
    padding: 12px 16px;
    margin: 4px 2px;
}

.card-title {
    font-weight: bold;
    color: #3498db;
    margin-bottom: 4px;
}

.field-label {
    color: alpha(currentColor, 0.75);
    font-weight: 600;
}

.mono-font {
    font-family: monospace;
}

.pill-box {
    background-color: rgba(128, 128, 128, 0.12);
    border-radius: 8px;
    padding: 6px 10px;
}

progressbar {
    min-width: 200px;
}

progressbar trough {
    min-height: 12px;
    border-radius: 6px;
    background-color: rgba(128, 128, 128, 0.16);
}

progressbar progress {
    border-radius: 6px;
}

progressbar.progress-low progress {
    background-color: #27ae60;
    background-image: none;
}

progressbar.progress-medium progress {
    background-color: #f39c12;
    background-image: none;
}

progressbar.progress-high progress {
    background-color: #e74c3c;
    background-image: none;
}

.footer-text {
    font-size: 11px;
    color: alpha(currentColor, 0.55);
}
"""


def apply_custom_css():
    """Применение стилей для современного внешнего вида интерфейса."""
    try:
        css_provider = Gtk.CssProvider()
        css_provider.load_from_data(CSS_DATA.encode("utf-8"))
        screen = Gdk.Screen.get_default()
        if screen:
            Gtk.StyleContext.add_provider_for_screen(
                screen, css_provider, Gtk.STYLE_PROVIDER_PRIORITY_APPLICATION
            )
    except Exception as e:
        print(f"Warning: could not load custom CSS: {e}")


class MainWindow(Gtk.Window):
    """
    Главное окно приложения System Monitor с современным карточным дизайном.
    """

    def __init__(self):
        super().__init__(title="System Monitor")
        self.set_default_size(640, 720)
        self.set_border_width(12)
        self.set_position(Gtk.WindowPosition.CENTER)

        apply_custom_css()

        version = get_version()
        self.is_updating = False

        # 1. Современный HeaderBar
        header_bar = Gtk.HeaderBar()
        header_bar.set_show_close_button(True)
        header_bar.set_title("System Monitor")
        header_bar.set_subtitle(f"v{version}")
        self.set_titlebar(header_bar)

        # Спиннер и кнопка обновления в HeaderBar
        self.spinner = Gtk.Spinner()
        header_bar.pack_end(self.spinner)

        self.refresh_btn = Gtk.Button(label="⟳ Refresh")
        self.refresh_btn.set_tooltip_text("Обновить данные системы сейчас")
        self.refresh_btn.connect("clicked", self.on_refresh_clicked)
        header_bar.pack_end(self.refresh_btn)

        # 2. Основной контейнер с прокруткой
        scrolled = Gtk.ScrolledWindow()
        scrolled.set_policy(Gtk.PolicyType.NEVER, Gtk.PolicyType.AUTOMATIC)
        self.add(scrolled)

        main_box = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=10)
        main_box.set_margin_left(6)
        main_box.set_margin_right(6)
        main_box.set_margin_top(6)
        main_box.set_margin_bottom(6)
        scrolled.add(main_box)

        self.labels = {}

        # КАРТОЧКА 1: 🖥️ Система
        card_sys, grid_sys = self._create_card("🖥️  Системная информация")
        main_box.pack_start(card_sys, False, False, 0)

        self._add_text_row(grid_sys, 0, "ОС и Ядро:", "os")
        self._add_text_row(grid_sys, 1, "Имя компьютера:", "hostname")
        self._add_text_row(grid_sys, 2, "Время работы (Uptime):", "uptime")

        # КАРТОЧКА 2: ⚡ Производительность (CPU и RAM)
        card_perf, grid_perf = self._create_card("⚡  Производительность")
        main_box.pack_start(card_perf, False, False, 0)

        # CPU Row
        cpu_title = self._create_field_label("Процессор (CPU):")
        grid_perf.attach(cpu_title, 0, 0, 1, 1)

        cpu_box = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=3)
        cpu_box.set_hexpand(True)
        self.cpu_bar = Gtk.ProgressBar()
        self.cpu_bar.set_show_text(True)
        self.cpu_label = Gtk.Label(xalign=0, halign=Gtk.Align.START)
        self.cpu_label.get_style_context().add_class("footer-text")
        cpu_box.pack_start(self.cpu_bar, False, False, 0)
        cpu_box.pack_start(self.cpu_label, False, False, 0)
        grid_perf.attach(cpu_box, 1, 0, 1, 1)

        # RAM Row
        ram_title = self._create_field_label("Память (RAM):")
        grid_perf.attach(ram_title, 0, 1, 1, 1)

        ram_box = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=3)
        ram_box.set_hexpand(True)
        self.ram_bar = Gtk.ProgressBar()
        self.ram_bar.set_show_text(True)
        self.ram_label = Gtk.Label(xalign=0, halign=Gtk.Align.START)
        self.ram_label.get_style_context().add_class("footer-text")
        ram_box.pack_start(self.ram_bar, False, False, 0)
        ram_box.pack_start(self.ram_label, False, False, 0)
        grid_perf.attach(ram_box, 1, 1, 1, 1)

        # КАРТОЧКА 3: 💾 Накопители
        card_disk, grid_disk = self._create_card("💾  Дисковые накопители")
        main_box.pack_start(card_disk, False, False, 0)

        self.disks_container = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=8)
        self.disks_container.set_hexpand(True)
        grid_disk.attach(self.disks_container, 0, 0, 2, 1)

        # КАРТОЧКА 4: 🌐 Сеть и Трафик
        card_net, grid_net = self._create_card("🌐  Сеть и Трафик")
        main_box.pack_start(card_net, False, False, 0)

        self._add_text_row(grid_net, 0, "IP-адрес:", "ip", mono=True)
        self._add_text_row(grid_net, 1, "DNS-серверы:", "dns", mono=True)

        # Трафик pill-box
        net_speed_title = self._create_field_label("Скорость трафика:")
        grid_net.attach(net_speed_title, 0, 2, 1, 1)

        speed_pill = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL)
        speed_pill.get_style_context().add_class("pill-box")
        self.labels["net_speed"] = Gtk.Label(xalign=0, halign=Gtk.Align.START)
        self.labels["net_speed"].set_selectable(True)
        speed_pill.pack_start(self.labels["net_speed"], True, True, 0)
        grid_net.attach(speed_pill, 1, 2, 1, 1)

        # КАРТОЧКА 5: 🐳 Docker
        card_docker, grid_docker = self._create_card("🐳  Служба Docker")
        main_box.pack_start(card_docker, False, False, 0)

        self._add_text_row(grid_docker, 0, "Статус службы:", "docker_status")
        self._add_text_row(grid_docker, 1, "Контейнеры:", "docker_containers")
        self._add_text_row(grid_docker, 2, "Образы (Images):", "docker_images")

        # Подвал: время обновления и статус
        footer_box = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=10)
        footer_box.set_margin_top(4)
        footer_box.set_margin_bottom(2)

        self.status_label = Gtk.Label(xalign=0, halign=Gtk.Align.START)
        self.status_label.get_style_context().add_class("footer-text")
        footer_box.pack_start(self.status_label, False, False, 0)

        self.labels["time"] = Gtk.Label(xalign=1, halign=Gtk.Align.END)
        self.labels["time"].get_style_context().add_class("footer-text")
        footer_box.pack_end(self.labels["time"], False, False, 0)

        main_box.pack_start(footer_box, False, False, 0)

        # Первичное обновление
        self.trigger_async_update()

        # Регулярный фоновый опрос каждые 3 секунды
        GLib.timeout_add_seconds(3, self.on_timer_tick)

    def _create_card(self, title: str) -> Tuple[Gtk.Box, Gtk.Grid]:
        """Создает карточку с заголовком и внутренней сеткой."""
        card = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=10)
        card.get_style_context().add_class("card")

        title_lbl = Gtk.Label(xalign=0, halign=Gtk.Align.START)
        title_lbl.set_markup(f"<b>{title}</b>")
        title_lbl.get_style_context().add_class("card-title")
        card.pack_start(title_lbl, False, False, 0)

        grid = Gtk.Grid()
        grid.set_column_spacing(20)
        grid.set_row_spacing(10)
        card.pack_start(grid, True, True, 0)

        return card, grid

    def _create_field_label(self, text: str) -> Gtk.Label:
        """Метка названия поля."""
        lbl = Gtk.Label(label=text, xalign=0, halign=Gtk.Align.START, valign=Gtk.Align.START)
        lbl.get_style_context().add_class("field-label")
        return lbl

    def _add_text_row(self, grid: Gtk.Grid, row: int, label_text: str, key: str, mono: bool = False):
        """Добавление строки с текстовой информацией в сетку карточки."""
        title = self._create_field_label(label_text)
        grid.attach(title, 0, row, 1, 1)

        val_label = Gtk.Label(xalign=0, halign=Gtk.Align.START, valign=Gtk.Align.START)
        val_label.set_line_wrap(True)
        val_label.set_selectable(True)
        if mono:
            val_label.get_style_context().add_class("mono-font")
        grid.attach(val_label, 1, row, 1, 1)
        self.labels[key] = val_label

    def _update_progress_bar_style(self, bar: Gtk.ProgressBar, percent: float, text: str):
        """Обновление прогресс-бара и цвета в зависимости от нагрузки."""
        fraction = max(0.0, min(1.0, percent / 100.0))
        bar.set_fraction(fraction)
        bar.set_text(text)

        context = bar.get_style_context()
        context.remove_class("progress-low")
        context.remove_class("progress-medium")
        context.remove_class("progress-high")

        if percent < 70.0:
            context.add_class("progress-low")
        elif percent < 90.0:
            context.add_class("progress-medium")
        else:
            context.add_class("progress-high")

    def on_timer_tick(self) -> bool:
        """Таймер регулярного опроса."""
        self.trigger_async_update()
        return True

    def on_refresh_clicked(self, button: Gtk.Button):
        """Обработчик нажатия на кнопку ручного обновления."""
        self.trigger_async_update()

    def trigger_async_update(self):
        """Запуск сбора системных данных в фоновом потоке."""
        if self.is_updating:
            return

        self.is_updating = True
        self.refresh_btn.set_sensitive(False)
        self.spinner.start()
        self.status_label.set_text("Обновление данных...")

        worker = threading.Thread(target=self._collect_data_worker, daemon=True)
        worker.start()

    def _collect_data_worker(self):
        """Сбор данных в отдельном потоке."""
        try:
            hw = get_hardware_summary()
            net_speed = get_network_speed()
            docker_info = get_docker_summary()
            data = {
                "hostname": get_hostname(),
                "os": get_os_info(),
                "uptime": get_uptime(),
                "ip": get_default_ip(),
                "dns": get_dns_servers(),
                "net_speed": net_speed,
                "hw": hw,
                "docker": docker_info,
                "time": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
                "status": "● Онлайн",
                "error": None
            }
        except Exception as e:
            data = {
                "status": f"✗ Ошибка: {str(e)[:40]}",
                "error": str(e)
            }

        GLib.idle_add(self._apply_collected_data, data)

    def _apply_collected_data(self, data: dict):
        """Обновление интерфейса в главном потоке GTK."""
        try:
            if not data.get("error"):
                # 1. Текстовые поля
                for key in ["hostname", "os", "uptime", "ip", "dns"]:
                    if key in self.labels and key in data:
                        self.labels[key].set_text(str(data[key]))

                # 2. Сетевой трафик
                net_speed = data.get("net_speed", {})
                rx_t = net_speed.get("rx_text", "0 B/s")
                tx_t = net_speed.get("tx_text", "0 B/s")
                if "net_speed" in self.labels:
                    self.labels["net_speed"].set_markup(
                        f"<span foreground='#27ae60'>↓</span> <b>{rx_t}</b>    "
                        f"<span foreground='#2980b9'>↑</span> <b>{tx_t}</b>"
                    )

                # 3. Аппаратные метрики
                hw = data.get("hw", {})

                # CPU
                cpu = hw.get("cpu", {})
                cpu_pct = float(cpu.get("percent", 0.0))
                cpu_cores = cpu.get("cores", 1)
                self._update_progress_bar_style(
                    self.cpu_bar, cpu_pct, f"{cpu_pct:.1f}%"
                )
                self.cpu_label.set_text(f"Количество ядер: {cpu_cores}")

                # RAM
                ram = hw.get("ram", {})
                ram_pct = float(ram.get("percent", 0.0))
                used_gb = ram.get("used_gb", 0.0)
                total_gb = ram.get("total_gb", 0.0)
                avail_gb = ram.get("available_gb", 0.0)
                self._update_progress_bar_style(
                    self.ram_bar, ram_pct, f"{ram_pct:.1f}%"
                )
                self.ram_label.set_text(
                    f"Использовано: {used_gb:.1f} GB из {total_gb:.1f} GB (Свободно: {avail_gb:.1f} GB)"
                )

                # Диски
                disks = hw.get("disks", [])
                for child in self.disks_container.get_children():
                    self.disks_container.remove(child)

                if disks:
                    for disk in disks[:6]:
                        mp = disk.get("mountpoint", "/")
                        pct = float(disk.get("percent", 0.0))
                        d_used = disk.get("used_gb", 0.0)
                        d_total = disk.get("total_gb", 0.0)

                        item_box = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=3)
                        item_box.set_hexpand(True)

                        info_label = Gtk.Label(xalign=0, halign=Gtk.Align.START)
                        info_label.set_markup(
                            f"<b>{mp}</b>  <small style='color: alpha(currentColor, 0.7);'>({d_used:.1f} / {d_total:.1f} GB)</small>"
                        )
                        item_box.pack_start(info_label, False, False, 0)

                        bar = Gtk.ProgressBar()
                        bar.set_show_text(True)
                        bar.set_hexpand(True)
                        self._update_progress_bar_style(bar, pct, f"{pct:.1f}%")
                        item_box.pack_start(bar, False, False, 0)

                        self.disks_container.pack_start(item_box, False, False, 0)
                else:
                    empty_label = Gtk.Label(label="Информация о дисках недоступна", xalign=0)
                    self.disks_container.pack_start(empty_label, False, False, 0)

                self.disks_container.show_all()

                # 4. Данные Docker
                docker = data.get("docker", {})
                if "docker_status" in self.labels:
                    self.labels["docker_status"].set_text(docker.get("status_display", "—"))
                if "docker_containers" in self.labels:
                    if docker.get("status") == "running":
                        self.labels["docker_containers"].set_text(docker.get("text", "—"))
                    else:
                        self.labels["docker_containers"].set_text(docker.get("text", "—"))
                if "docker_images" in self.labels:
                    img_count = docker.get("images_count", 0)
                    if docker.get("status") == "running":
                        self.labels["docker_images"].set_text(f"{img_count} образов")
                    else:
                        self.labels["docker_images"].set_text("—")

                # Время обновления
                if "time" in self.labels:
                    self.labels["time"].set_text(f"Обновлено: {data.get('time', '')}")

            self.status_label.set_text(data.get("status", "Готово"))
        finally:
            self.spinner.stop()
            self.refresh_btn.set_sensitive(True)
            self.is_updating = False

        return False
