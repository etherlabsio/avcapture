#!/bin/sh

cd /tmp
wget $FFMPEG_TGZ_URI &&  tar -xzf ubuntu16.04.tgz && mv bin/ffmpeg /usr/local/bin && mv lib/* /usr/local/lib/ && rm ubuntu16.04.tgz && /bin/server