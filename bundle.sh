#!/bin/bash
zip -r -j app_contents.zip app_contents/*
mkdir __items
cp ./app_contents/.ironfist.json ./__items/.ironfist.json
cp ./app_contents.zip ./__items/app_contents.zip
go get github.com/gobuffalo/packr
cd client
if hash packr 2>/dev/null; then
    packr build -o ../bundled_application
else
    ~/go/bin/packr build -o ../bundled_application
fi
cd ..
rm -rf __items
rm ./app_contents.zip
