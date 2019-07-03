#!/bin/bash
service probebot stop
cp probebot-initd /etc/init.d/probebot
insserv probebot
service probebot start
