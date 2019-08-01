#!/bin/sh

WORK_DIR=/data
cd $WORK_DIR

if [ ! -f $WORK_DIR/ffmpeg.tgz ] || [ ! -z $DISABLE_FFMPEG_CACHE ]; then
	wget $FFMPEG_TGZ_URI -O ffmpeg.tgz
fi

if [ ! -f $WORK_DIR/ffmpeg.tgz ]; then
	echo "failed to fetch ffmpeg binary"
	exit
fi

tar -xzf ffmpeg.tgz && mv ffmpeg /usr/local/bin/

if [ ! -z $FFMPEG_DEPS_URI ]; then
	if [ ! -f $WORK_DIR/ffmpeg-deps.tgz ] || [ ! -z $DISABLE_FFMPEG_CACHE ]; then
		wget $FFMPEG_DEPS_URI -O ffmpeg-deps.tgz
	fi
	if [ ! -f $WORK_DIR/ffmpeg-deps.tgz ]; then
		echo "failed to fetch ffmpeg dependencies"
		exit
	fi
	tar -xzf ffmpeg-deps.tgz && sudo mv lib/* /usr/local/lib/
fi

echo "ffmpeg setup successful"
