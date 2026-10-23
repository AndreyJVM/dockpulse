"""
main_window.py - Графический интерфейс на GTK 3 с нативными цветными ProgressBar.
"""

import threading
from datetime import datetime

import gi
gi.require_version('Gtk', '3.0')
from gi.repository import Gtk, GLib, Gdk

from monitors.system import get_os_info, get_uptime, get_hostname
from monitors.network import get_default_ip, get_dns_servers
from monitors.hardware import get_hardware_summary


CSS_DATA = """
progressbar {
    min-width: 260px;
}
progressbar trough {
    min-height: 16px;
    border-radius: 8px;
    background-color: rgba(128, 128, 128, 0.2);
}
progressbar progress {
    border-radius: 8px;
}
progressbar.progress-low progress {
    background-color: #27ae60;
    background-image: none;
}
progressbar.progress-medium progress {
    background-color: #e67e22;
    background-image: none;
}
progressbar.progress-high progress {
    background-color: #e74c3c;
    background-image: none;
}
"""


def apply_custom_css():
    """Применение стилей для нативных цветных прогресс-баров."""
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
    Главное окно приложения System Monitor с нативными прогресс-барами.
    """

    def __init__(self):
        super().__init__(title="System Monitor")
        self.set_default_size(620, 560)
        self.set_border_width(14)
        self.set_position(Gtk.WindowPosition.CENTER)

        apply_custom_css()

        self.is_updating = False

        # Основной вертикальный контейнер
        vbox = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=12)
        self.add(vbox)

        # Заголовок
        header = Gtk.Label()
        header.set_markup("<big><b>System Information Monitor</b></big>")
        vbox.pack_start(header, False, False, 0)

        # ScrolledWindow для прокрутки при большом количестве дисков
        scrolled = Gtk.ScrolledWindow()
        scrolled.set_policy(Gtk.PolicyType.AUTOMATIC, Gtk.PolicyType.AUTOMATIC)
        scrolled.set_shadow_type(Gtk.ShadowType.ETCHED_IN)
        vbox.pack_start(scrolled, True, True, 0)

        # Сетка параметров
        grid = Gtk.Grid()
        grid.set_column_spacing(18)
        grid.set_row_spacing(14)
        grid.set_margin_top(14)
        grid.set_margin_bottom(14)
        grid.set_margin_left(14)
        grid.set_margin_right(14)
        scrolled.add(grid)

        self.labels = {}
        row = 0

        # Текстовые информационные поля
        text_fields = [
            ("Hostname:", "hostname"),
            ("Operating System:", "os"),
            ("Uptime:", "uptime"),
            ("IP Address:", "ip"),
            ("DNS Servers:", "dns"),
        ]

        for label_text, key in text_fields:
            title = self._create_section_label(label_text)
            grid.attach(title, 0, row, 1, 1)

            val_label = Gtk.Label(xalign=0, halign=Gtk.Align.START)
            val_label.set_line_wrap(True)
            val_label.set_selectable(True)
            grid.attach(val_label, 1, row, 1, 1)
            self.labels[key] = val_label
            row += 1

        # 1. CPU Usage с нативным ProgressBar
        title_cpu = self._create_section_label("CPU Usage:")
        grid.attach(title_cpu, 0, row, 1, 1)

        cpu_box = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=4)
        cpu_box.set_hexpand(True)
        self.cpu_bar = Gtk.ProgressBar()
        self.cpu_bar.set_show_text(True)
        self.cpu_label = Gtk.Label(xalign=0, halign=Gtk.Align.START)
        cpu_box.pack_start(self.cpu_bar, False, False, 0)
        cpu_box.pack_start(self.cpu_label, False, False, 0)
        grid.attach(cpu_box, 1, row, 1, 1)
        row += 1

        # 2. RAM Usage с нативным ProgressBar
        title_ram = self._create_section_label("RAM Usage:")
        grid.attach(title_ram, 0, row, 1, 1)

        ram_box = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=4)
        ram_box.set_hexpand(True)
        self.ram_bar = Gtk.ProgressBar()
        self.ram_bar.set_show_text(True)
        self.ram_label = Gtk.Label(xalign=0, halign=Gtk.Align.START)
        ram_box.pack_start(self.ram_bar, False, False, 0)
        ram_box.pack_start(self.ram_label, False, False, 0)
        grid.attach(ram_box, 1, row, 1, 1)
        row += 1

        # 3. Disk Usage с динамическими ProgressBar для каждого раздела
        title_disk = self._create_section_label("Disk Usage:")
        grid.attach(title_disk, 0, row, 1, 1)

        self.disks_container = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=8)
        self.disks_container.set_hexpand(True)
        grid.attach(self.disks_container, 1, row, 1, 1)
        row += 1

        # Время последнего обновления
        title_time = self._create_section_label("Last Update:")
        grid.attach(title_time, 0, row, 1, 1)
        self.labels["time"] = Gtk.Label(xalign=0, halign=Gtk.Align.START)
        self.labels["time"].set_selectable(True)
        grid.attach(self.labels["time"], 1, row, 1, 1)

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

    def _create_section_label(self, text: str) -> Gtk.Label:
        """Создание заголовка параметра."""
        lbl = Gtk.Label()
        lbl.set_markup(f"<b>{text}</b>")
        lbl.set_xalign(0)
        lbl.set_halign(Gtk.Align.START)
        lbl.set_valign(Gtk.Align.START)
        return lbl

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
        self.refresh_btn.set_label("Updating...")
        self.statusbar.push(self.status_context, "Fetching system data...")

        worker = threading.Thread(target=self._collect_data_worker, daemon=True)
        worker.start()

    def _collect_data_worker(self):
        """Сбор данных в отдельном потоке."""
        try:
            hw = get_hardware_summary()
            data = {
                "hostname": get_hostname(),
                "os": get_os_info(),
                "uptime": get_uptime(),
                "ip": get_default_ip(),
                "dns": get_dns_servers(),
                "hw": hw,
                "time": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
                "status": "✓ Updated successfully",
                "error": None
            }
        except Exception as e:
            data = {
                "status": f"✗ Error: {str(e)[:40]}",
                "error": str(e)
            }

        GLib.idle_add(self._apply_collected_data, data)

    def _apply_collected_data(self, data: dict):
        """Обновление интерфейса в главном потоке GTK."""
        try:
            if not data.get("error"):
                # 1. Текстовые поля
                for key in ["hostname", "os", "uptime", "ip", "dns", "time"]:
                    if key in self.labels and key in data:
                        self.labels[key].set_text(str(data[key]))

                # 2. Аппаратные метрики
                hw = data.get("hw", {})

                # CPU
                cpu = hw.get("cpu", {})
                cpu_pct = float(cpu.get("percent", 0.0))
                cpu_cores = cpu.get("cores", 1)
                self._update_progress_bar_style(
                    self.cpu_bar, cpu_pct, f"{cpu_pct:.1f}%"
                )
                self.cpu_label.set_text(f"Cores: {cpu_cores}")

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
                    f"Used: {used_gb:.1f} GB / Total: {total_gb:.1f} GB (Available: {avail_gb:.1f} GB)"
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

                        item_box = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=2)
                        item_box.set_hexpand(True)

                        info_label = Gtk.Label(xalign=0, halign=Gtk.Align.START)
                        info_label.set_markup(
                            f"<b>{mp}</b>  <small>({d_used:.1f} / {d_total:.1f} GB)</small>"
                        )
                        item_box.pack_start(info_label, False, False, 0)

                        bar = Gtk.ProgressBar()
                        bar.set_show_text(True)
                        bar.set_hexpand(True)
                        self._update_progress_bar_style(bar, pct, f"{pct:.1f}%")
                        item_box.pack_start(bar, False, False, 0)

                        self.disks_container.pack_start(item_box, False, False, 0)
                else:
                    empty_label = Gtk.Label(label="No disk information available", xalign=0)
                    self.disks_container.pack_start(empty_label, False, False, 0)

                self.disks_container.show_all()

            self.statusbar.push(self.status_context, data.get("status", "Updated"))
        finally:
            self.refresh_btn.set_label("Refresh Now")
            self.refresh_btn.set_sensitive(True)
            self.is_updating = False

        return False
