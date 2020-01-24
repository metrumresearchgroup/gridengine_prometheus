#!/bin/bash

SERVICE=gridengine_prometheus
SERVICE_INSTALLED_MESSAGE="Service is already installed. No further action required."

#Is this a systemd service?

if systemctl > /dev/null 2>&1 ;
then
    echo "This is a SystemD server"

    #Unit file sill already exist as we'll put it there.
    #We need to see if the unit is registered and loaded with SystemD

    if systemctl status ${SERVICE}> /dev/null 2>&1
    then
      echo "$SERVICE_INSTALLED_MESSAGE"
      exit 0
    fi

    #Doesn't look like it's installed. Let's do the needful
    systemctl enable $SERVICE
    systemctl daemon-reload
    if systemctl start $SERVICE
    then
        echo "Successfully setup and started SystemD Unit"
        exit 0
    fi
fi

if ! systemctl > /dev/null 2>&1 ;
then
  echo "This is an initscript / init D server"

  #First we need to make the init script executable
  chmod +x /etc/init.d/${SERVICE}

  #Check to see if service is enabled
  SERVICE_PRESENCE_COUNT=`service --status-all 2>&1 | grep ${SERVICE} | wc -l`

  if [ "$SERVICE_PRESENCE_COUNT" -gt 0 ] ;
  then
      #Already appears to be installed
      echo "$SERVICE_INSTALLED_MESSAGE"
      exit 0
  fi

  #Doesn't look like it's installed. Let's do the needful
  update-rc.d ${SERVICE} defaults

  if update-rc.d ${SERVICE} enable
  then
      echo "Successfully installed SystemV Init script"
    exit 0
  fi
fi