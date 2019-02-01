#!/bin/bash

ffmpeg -i video.webm -vf scale=192:96 -vframes 10000 /tmp/thumb%05d.bmp
