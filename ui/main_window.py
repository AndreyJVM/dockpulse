"""
main_window.py - Графический интерфейс на GTK 3 с фоновым обновлением данных.
"""

import threading
from datetime import datetime

import gi
gi.require_version('Gtk', '3.0')
from gi.repository import Gtk, GLib

from monitors.system import get_os_info, get_uptime, get_hostname
from monitors.network import get_default_ip, get_dns_servers
from monitors.hardware import get_cpu_info, get_ram_info, get_hardware_summary


class MainWindow(Gtk.Window):
    """
    Главное окно приложения System Monitor.
    """

    def __init__(self):
        super().__init__(title="System Monitor")
        self.set_default_size(580, 520)
        self.set_border_width(12)
        self.set_position(Gtk.WindowPosition.CENTER)

        self.is_updating = False

        # Основной вертикальный контейнер
        vbox = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=12)
        self.add(vbox)

        # Заголовок
        header = Gtk.Label()
        header.set_markup("<big><b>System Information Monitor</b></big>")
        vbox.pack_start(header, False, False, 0)

        # ScrolledWindow для прокрутки при большом количестве разделов диска
        scrolled = Gtk.ScrolledWindow()
        scrolled.set_policy(Gtk.PolicyType.AUTOMATIC, Gtk.PolicyType.AUTOMATIC)
        scrolled.set_shadow_type(Gtk.ShadowType.ETCHED_IN)
        vbox.pack_start(scrolled, True, True, 0)

        # Сетка параметров
        grid = Gtk.Grid()
        grid.set_column_spacing(15)
        grid.set_row_spacing(12)
        grid.set_margin_top(12)
        grid.set_margin_bottom(12)
        grid.set_margin_left(12)
        grid.set_margin_right(12)
        scrolled.add(grid)

        # Определение элементов интерфейса
        self.labels = {}
        info_items = [
            ("Hostname:", "hostname"),
            ("Operating System:", "os"),
            ("Uptime:", "uptime"),
            ("IP Address:", "ip"),
            ("DNS Servers:", "dns"),
            ("CPU Usage:", "cpu"),
            ("RAM Usage:", "ram"),
            ("Disk Usage:", "disk"),
            ("Last Update:", "time"),
        ]

        for row, (label_text, key) in enumerate(info_items):
            title = Gtk.Label()
            title.set_markup(f"<b>{label_text}</b>")
            title.set_xalign(0)
            title.set_halign(Gtk.Align.START)
            title.set_valign(Gtk.Align.START)
            grid.attach(title, 0, row, 1, 1)

            value = Gtk.Label()
            value.set_xalign(0)
            value.set_halign(Gtk.Align.START)
            value.set_line_wrap(True)
            value.set_selectable(True)
            grid.attach(value, 1, row, 1, 1)

            self.labels[key] = value

        # Панель управления и кнопка обновления
        button_box = Gtk.Box(spacing=10)
        self.refresh_btn = Gtk.Button(label="Refresh Now")
        self.refresh_btn.connect("clicked", self.on_refresh_clicked)
        button_box.pack_end(self.refresh_btn, False, False, 0)
        vbox.pack_start(button_box, False, False, 0)

        # Статус-бар
        self.statusbar = Gtk.Statusbar()
        self.status_context = self.statusbar.get_context_id("status")
        vbox.pack_start(self.statusbar, False, False, 0)

        # Первичное обновление данных
        self.trigger_async_update()

        # Автоматическое обновление каждые 10 секунд
        GLib.timeout_add_seconds(10, self.on_timer_tick)

    def on_timer_tick(self) -> bool:
        """Таймер регулярного опроса."""
        self.trigger_async_update()
        return True

    def on_refresh_clicked(self, button: Gtk.Button):
        """Обработчик нажатия на кнопку ручного обновления."""
        self.trigger_async_update()

    def trigger_async_update(self):
        """
        Запуск сбора системных данных в фоновом потоке,
        чтобы не блокировать графический интерфейс GTK.
        """
        if self.is_updating:
            return

        self.is_updating = True
        self.refresh_btn.set_sensitive(False)
        self.refresh_btn.set_label("Updating...")
        self.statusbar.push(self.status_context, "Fetching system data...")

        worker = threading.Thread(target=self._collect_data_worker, daemon=True)
        worker.start()

    def _collect_data_worker(self):
        """Сбор данных в отдельном потоке (I/O, сокеты, системные вызовы)."""
        try:
            hw = get_hardware_summary()
            data = {
                "hostname": get_hostname(),
                "os": get_os_info(),
                "uptime": get_uptime(),
                "ip": get_default_ip(),
                "dns": get_dns_servers(),
                "cpu": hw["cpu"]["text"],
                "ram": hw["ram"]["text"],
                "disk": hw["disks_text"],
                "time": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
                "status": "✓ Updated successfully",
                "error": None
            }
        except Exception as e:
            data = {
                "status": f"✗ Error: {str(e)[:40]}",
                "error": str(e)
            }

        # Безопасная передача данных в поток отрисовки GTK
        GLib.idle_add(self._apply_collected_data, data)

    def _apply_collected_data(self, data: dict):
        """Обновление интерфейса в главном потоке GTK."""
        try:
            if not data.get("error"):
                for key in ["hostname", "os", "uptime", "ip", "dns", "cpu", "ram", "disk", "time"]:
                    if key in self.labels and key in data:
                        self.labels[key].set_text(str(data[key]))

            self.statusbar.push(self.status_context, data.get("status", "Updated"))
        finally:
            self.refresh_btn.set_label("Refresh Now")
            self.refresh_btn.set_sensitive(True)
            self.is_updating = False

        return False
