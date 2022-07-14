#!/bin/sh
BEFORE_UPGRADE_FILE="/tmp/gipgee-apk-list-before-upgrade.txt"
AFTER_UPGRADE_FILE="/tmp/gipgee-apk-list-after-upgrade.txt"
if [ -z "$1" ]; then
	echo "First argument must be the path of the result file"
	exit 1
fi
echo "Checking for upgrades, writing result to file $1"
apk info -v | sort > "$BEFORE_UPGRADE_FILE"
apk upgrade --no-cache
apk info -v | sort > "$AFTER_UPGRADE_FILE"
echo "Diffing before and after upgrade"
diff "$BEFORE_UPGRADE_FILE" "$AFTER_UPGRADE_FILE"
RET=$?
if [ "$RET" -eq 0 ]; then
	echo "unchanged" > "$1"
elif [ "$RET" -eq 1 ]; then
	echo "changed" > "$1"
else
	echo "Unexpected diff exit code $RET"
	exit 1
fi
