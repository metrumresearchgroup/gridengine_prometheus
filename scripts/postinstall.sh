#!/bin/bash

SERVICE=gridengine_prometheus

systemctl enable $SERVICE
systemctl daemon-reload
systemctl start $SERVICE