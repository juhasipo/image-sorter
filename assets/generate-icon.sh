#!/bin/bash

set -eu

inkscape -z -w 32 -h 32 icon.svg -e icon-32x32.png
inkscape -z -w 64 -h 64 icon.svg -e icon-64x64.png
inkscape -z -w 128 -h 128 icon.svg -e icon-128x128.png
