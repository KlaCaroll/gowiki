#! /usr/bin/env bash

while true
do 
	fswatch --one-event *.go static/*.css tmpl/*.html > /dev/null
	pkill -9 wiki
	go build -o wiki wiki.go
	./wiki &
done
