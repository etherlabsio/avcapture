#!/bin/sh

WORK_DIR=/data
cd $WORK_DIR

if [ -z $UPDATE_FFMPEG_CACHE ]
then
    UPDATE_FFMPEG_CACHE="1"
fi 

if [ ! -f $WORK_DIR/ffmpeg.tgz ] || [ $UPDATE_FFMPEG_CACHE -eq "1" ]
then 
    wget $FFMPEG_TGZ_URI -O ffmpeg.tgz
    wget $FFMPEG_DEPS_URI -O ffmpeg-deps.tgz
fi

if [ ! -f $WORK_DIR/ffmpeg.tgz ]
then 
    echo "failed to fetch ffmpeg binary"
    exit 
fi
if [ ! -f $WORK_DIR/ffmpeg-deps.tgz ]
then 
    echo "failed to fetch ffmpeg dependencies"
    exit 
fi

tar -xzf ffmpeg.tgz && mv ffmpeg /usr/local/bin
tar -xzf ffmpeg-deps.tgz &&  mv lib/* /usr/local/lib/ 
/bin/server