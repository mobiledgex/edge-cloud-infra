#!/bin/bash

memKB=$(grep MemTotal /proc/meminfo | awk '{ print $2}')

vcpus=$(grep -c processor /proc/cpuinfo)

diskGB=$(fdisk -l | head -n 1 | awk '{ print $3}')

echo "$memKB,$vcpus,$diskGB"
