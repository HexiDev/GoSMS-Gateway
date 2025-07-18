#!/bin/sh
# Build GoSMS Gateway for multiple architectures
set -e

BIN=send-sms
SRC=main.go
OUT=out
# UPX executable for compression (expected at ./upx.exe)
UPX=./upx.exe

mkdir -p $OUT

# Linux ARM64 (OpenWRT/GL.iNet X3000)
echo "Building for linux/arm64..."
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $OUT/${BIN}-arm64 $SRC &

# Linux AMD64 (x86_64)
echo "Building for linux/amd64..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $OUT/${BIN}-amd64 $SRC &

# Windows AMD64
echo "Building for windows/amd64..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $OUT/${BIN}-win64.exe $SRC &

# Windows ARM64
echo "Building for windows/arm64..."
GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o $OUT/${BIN}-win-arm64.exe $SRC &


wait
echo "Builds complete. Binaries are in the '$OUT' directory."

if [ -f "$UPX" ]; then
  echo "Compressing all binaries with UPX in parallel..."
  for f in $OUT/*; do
    $UPX --best "$f" &
  done
  wait
  echo "UPX compression complete."
else
  echo "UPX not found, skipping compression."
fi
echo "Build process finished successfully."
read -p "Press any key to continue..." -n1 -s
