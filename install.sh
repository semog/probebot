#!/bin/bash

if [[ `systemctl` =~ -\.mount ]]; then
	sudo ./install-systemd.sh $*
elif [[ `/sbin/init --version` =~ upstart ]]; then
	echo Upstart system not supported. Attempting init.d install...
	sudo ./install-initd.sh $*
elif [[ -f /etc/init.d/cron && ! -h /etc/init.d/cron ]]; then
	sudo ./install-initd.sh $*
else
	echo Unknown init system
fi
