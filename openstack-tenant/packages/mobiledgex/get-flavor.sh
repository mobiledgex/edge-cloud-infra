#!/bin/bash

function isNum() {
  re='^[0-9]+$'
  if [[ ! $1 =~ $re ]] ; then
    return 1
  fi
  return 0
}

# Total RAM
totalMemMB=0
for memSizeMB in $(dmidecode -t memory | grep "Size:" | awk '{print $2}'); do
  if isNum $memSizeMB; then
    let totalMemMB+=$memSizeMB
  fi
done

# Number of vCPUs
vcpus=$(grep -c processor /proc/cpuinfo)

# Total disk size
totalDiskSectors=0
for block in $(ls -l /sys/block/ | grep -v "devices/virtual" | awk '{print $9}'); do
  if [[ ! -z $block ]]; then
    echo $block
    if [[ -f /sys/block/$block/size ]]; then
      sectors=$(cat /sys/block/$block/size)
      if isNum $sectors; then
        let totalDiskSectors+=$sectors
      fi
    fi
  fi
done
# assume sector size is always 512 bytes
let diskGB=$totalDiskSectors*512/1024/1024/1024

# Report system resource info
echo "$totalMemMB,$vcpus,$diskGB"
