#!/bin/bash
### BEGIN INIT INFO
# Provides:          skeleton
# Required-Start:    $remote_fs $syslog
# Required-Stop:     $remote_fs $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: Example initscript
# Description:       This file should be used to construct scripts to be
#                    placed in /etc/init.d.
### END INIT INFO

# Author: Darrell Breeden <dbreeden@thinq.com>
#
# Please remove the "Author" lines above and replace them
# with your own name if you copy and modify this script.

# Assumes Binary is placed in /usr/local/bin with all other custom binaries for opensips

# Do NOT "set -e"

PATH=/sbin:/usr/sbin:/bin:/usr/bin:/usr/local/bin:/opt/sge/bin/lx-amd64
DESC="Sun Grid Engine exporter for Prometheus"
NAME=gridengine_prometheus
CONFIG=/etc/gridengine_prometheus/config.yaml
DAEMON=/usr/local/bin/$NAME
DAEMON_ARGS="--config ${CONFIG}"
PIDFILE=/var/run/$NAME.pid
SCRIPTNAME=/etc/init.d/$NAME

# Exit if the package is not installed
[ -x "$DAEMON" ] || exit 0

# Read configuration variable file if it is present
[ -r /etc/default/$NAME ] && . /etc/default/$NAME

# Load the VERBOSE setting and other rcS variables
. /lib/init/vars.sh

# Define LSB log_* functions.
# Depend on lsb-base (>= 3.2-14) to ensure that this file is present
# and status_of_proc is working.
. /lib/lsb/init-functions

#
# Function that starts the daemon/service
#
do_start()
{

	echo "Starting SGE Exporter on port $PORT"
	# Return
	#   0 if daemon has been started
	#   1 if daemon was already running
	#   2 if daemon could not be started
	start-stop-daemon --start --quiet --pidfile $PIDFILE --exec $DAEMON --test > /dev/null \
		|| return 1
	start-stop-daemon --start -b --quiet --pidfile $PIDFILE --exec $DAEMON -- \
		$DAEMON_ARGS \
		|| return 2
	# Add code here, if necessary, that waits for the process to be ready
	# to handle requests from services started subsequently which depend
	# on this one.  As a last resort, sleep for some time.
	PID=`ps aux | grep ${DAEMON} | head -1 | awk '{print $2}'`
	echo $PID > $PIDFILE
}

#
# Function that stops the daemon/service
#
do_stop()
{
	echo "Killing PID `cat $PIDFILE`"
	# Return
	#   0 if daemon has been stopped
	#   1 if daemon was already stopped
	#   2 if daemon could not be stopped
	#   other if a failure occurred
	PID=`cat $PIDFILE`
	kill $PID

	PID_LINES=`ps aux | grep ${DAEMON} | grep -v "grep"| wc -l`

	if [ $PID_LINES -eq 0  ] ;
	then
		rm $PIDFILE
		return 0
	else
		echo "Couldn't kill process"
		return 2
	fi
}

do_status()
{
	if [ -e $PIDFILE ] ;
	then
		PID=`cat $PIDFILE`
		ps -p ${PID}
		PID_RETURN=$?
		if [ $PID_RETURN -eq 0 ] ;
		then
			echo "${NAME} is running. PID : ${PID}"
		else
			echo "${NAME} is not running"
		fi
	else
		echo "${NAME} is not running"
	fi
}

#
# Function that sends a SIGHUP to the daemon/service
#
do_reload() {
	#
	# If the daemon can reload its configuration without
	# restarting (for example, when it is sent a SIGHUP),
	# then implement that here.
	#
	start-stop-daemon --stop --signal 1 --quiet --pidfile $PIDFILE --name $NAME
	return 0
}

case "$1" in
  start)
	  [ "$VERBOSE" != no ] && log_daemon_msg "Starting $DESC" "$NAME"
    do_start
    case "$?" in
      0|1) [ "$VERBOSE" != no ] && log_end_msg 0 ;;
      2) [ "$VERBOSE" != no ] && log_end_msg 1 ;;
    esac
    ;;
  stop)
    [ "$VERBOSE" != no ] && log_daemon_msg "Stopping $DESC" "$NAME"
    do_stop
    case "$?" in
      0|1) [ "$VERBOSE" != no ] && log_end_msg 0 ;;
      2) [ "$VERBOSE" != no ] && log_end_msg 1 ;;
    esac
    ;;
  status)
     status_of_proc "$DAEMON" "$NAME" && exit 0 || exit $?
     ;;
  *)
    echo "Usage: $SCRIPTNAME {start|stop|status|restart}" >&2
    exit 3
	;;
esac

: