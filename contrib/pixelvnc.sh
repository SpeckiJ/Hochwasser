#!/bin/bash

# run vnc viewer in fullscreen on the selected monitor.
# restarts it on connectivity issues, press ctrl-c TWICE to quit.
#
# usage:      pixelvnc.sh [<host>[:<port>] [monitor-index]]
# depends on: tigervnc

host=${1:=vnc.schenklklopfer.de}
screen=${2:=1}

while :; do
  vncviewer ReconnectOnError=0 ViewOnly=1 FullScreen=1 FullScreenMode=Selected FullScreenSelectedMonitors=$screen $host
  sleep 1 # give room to ^C out of the loop
done

