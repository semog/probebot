#!/bin/bash
DATAFOLDER=/usr/local/share/appdata/probebot
APPFOLDER=/usr/local/lib/probebot
SYSTEMDFOLDER=/etc/systemd/system

if [ ! -e "probebot" ]; then
	echo "probebot binary not found. Please build it first."
	exit 1
fi

# If the service is being installed for the first time, then a bot
# token must be provided. If the service is being reinstalled, then
# the token is optional.
if [ -z "$1" ] && [ ! -e "$APPFOLDER/probebotsrv.sh" ]; then
	echo "Usage: $0 <bot token>"
	exit 1
fi

mkdir -p $DATAFOLDER/
mkdir -p $APPFOLDER/

systemctl --now disable probebot.service

cp probebot.service $SYSTEMDFOLDER/
cp probebot $APPFOLDER/

if [ ! -e "$APPFOLDER/probebotsrv.sh" ]; then
	cp probebotsrv.sh $APPFOLDER/
fi

if [ ! -z "$1" ]; then
	# Replace/update the bot token
	sed -i "s/BOTTOKEN=.*/BOTTOKEN=$1/g" $APPFOLDER/probebotsrv.sh
fi

# Turn off read-access to group/others to protect secret token
chmod go-r $APPFOLDER/probebotsrv.sh

systemctl --force enable probebot.service
systemctl start probebot.service
