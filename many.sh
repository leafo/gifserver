#!/bin/sh

function send() {
	curl http://localhost:9090/transcode?url=$1 > /dev/null & 
}

send leafo.net/dump/adit.gif
send leafo.net/dump/banana.gif
send leafo.net/dump/aditzoom.gif
send leafo.net/dump/hello.gif
send leafo.net/dump/exo1.gif
send leafo.net/dump/exo2.gif
