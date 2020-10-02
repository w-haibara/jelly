#!/bin/bash

zipName=portfolio
zipDirName=dist
dir=/var/www/html/home
dirName=home

if [ "$1" == "latest" ] || [ "$1" == "" ] || [ $# -gt 1 ]; then
	version=`curl https://github.com/w-haibara/portfolio/releases/latest | sed -e 's/<html><body>You are being <a href="https:\/\/github.com\/w-haibara\/portfolio\/releases\/tag\///' | sed -e 's/">redirected<\/a>.<\/body><\/html>//'`
	wget https://github.com/w-haibara/portfolio/releases/download/${version}/${zipName}.zip
else
	wget https://github.com/w-haibara/portfolio/releases/download/$1/${zipName}.zip
fi

rm -rf ${dir}/${dirName}
unzip portfolio.zip -d ${dir}
mv ${dir}/${zipDirName} ${dir}/${dirName}
rm ${zipName}.zip

