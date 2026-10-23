"""
app.py - Контроллер приложения System Monitor.
"""

import gi
gi.require_version('Gtk', '3.0')
from gi.repository import Gtk

from ui.main_window import MainWindow


class SystemMonitorApp:
    """
    Класс управления жизненным циклом приложения.
    """

    def __init__(self):
        self.window = MainWindow()
        self.window.connect("destroy", self.on_destroy)

    def on_destroy(self, window):
        Gtk.main_quit()

    def run(self):
        self.window.show_all()
        Gtk.main()
