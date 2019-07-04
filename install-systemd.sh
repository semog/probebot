#!/bin/bash
DATAFOLDER=/usr/local/share/appdata/probebot
APPFOLDER=/usr/local/lib/probebot
SYSTEMDFOLDER=/etc/systemd/system

mkdir -p $DATAFOLDER/
mkdir -p $APPFOLDER/

systemctl --now disable probebot.service

cp probebot.service $SYSTEMDFOLDER/
cp probebot $APPFOLDER/

if [ ! -e "$APPFOLDER/probebotsrv.sh" ]; then
	cp probebotsrv.sh $APPFOLDER/
fi

# Turn off read-access to group/others to protect secret token
chmod go-r $APPFOLDER/probebotsrv.sh

systemctl --force enable probebot.service
systemctl start probebot.service
