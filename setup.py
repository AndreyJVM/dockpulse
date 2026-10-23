#!/usr/bin/env python3
"""
setup.py - Конфигурация сборки и установки пакета.
"""

import os
from setuptools import setup, find_packages


def read_version():
    version_file = os.path.join(os.path.dirname(__file__), "VERSION")
    if os.path.exists(version_file):
        with open(version_file, "r", encoding="utf-8") as f:
            return f.read().strip()
    return "0.1.0"


setup(
    name="system-monitor",
    version=read_version(),
    description="Lightweight system information monitor written in Python and GTK 3",
    author="Andrey Vorobev",
    author_email="andrey.vorobev.aqa@gmail.com",
    packages=find_packages(),
    py_modules=["main", "app"],
    install_requires=[
        "psutil>=5.9.0",
        "PyGObject>=3.42.0",
    ],
    entry_points={
        "gui_scripts": [
            "system-monitor = main:main",
        ],
    },
    classifiers=[
        "Programming Language :: Python :: 3",
        "Operating System :: POSIX :: Linux",
        "Environment :: X11 Applications :: GTK",
        "Topic :: System :: Monitoring",
    ],
)
