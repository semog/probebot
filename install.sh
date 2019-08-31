#!/bin/bash
# Strip the debug symbols out of the executable.
go build -ldflags '-s'

# TODO: Test for init.d vs systemd.
sudo ./install-systemd.sh

# For init.d:
# sudo ./install-initd.sh

# TODO: Prompt for bot token and replace in the installed script file.

