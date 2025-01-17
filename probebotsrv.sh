#!/bin/bash
cd /usr/local/share/appdata/probebot/

while true; do
	/usr/local/lib/probebot/probebot -token=YOURTOKENHERE
	if [[ $? -eq 69 ]]; then
		echo "Probebot exited with code 69, restarting in 1 seconds..."
		sleep 1
	else
		echo "Probebot exited with code $?"
		break
	fi
done
