#!/bin/sh

export ETCDCTL_API=3
ENDPOINTS="http://mex-etcd-0.mex-etcd:2379,http://mex-etcd-1.mex-etcd:2379,http://mex-etcd-2.mex-etcd:2379"

echo "==== BEFORE COMPACTION ===="
etcdctl --endpoints="$ENDPOINTS" endpoint status -w table

REV=$( etcdctl --endpoints="$ENDPOINTS" get abc123 -w json | sed -n 's/.*"revision":\([0-9]*\).*/\1/p' )
echo "Current revision: $REV"

COMPREV=$(( REV - 5000 ))
echo "Compacting to $COMPREV"
OUT=$( etcdctl --endpoints="$ENDPOINTS" compact "$COMPREV" 2>&1; ); RC=$?

if [[ "$RC" != 0 ]]; then
        if echo "$OUT" | grep "required revision has been compacted" >/dev/null 2>&1; then
                echo "Compaction not necessary"
                exit 0
        fi

        # Unknown error during compaction; abort
        echo "$OUT" >&2
        exit "$RC"
fi

echo "Defragging"
etcdctl --endpoints="$ENDPOINTS" defrag

echo "==== AFTER COMPACTION ===="
etcdctl --endpoints="$ENDPOINTS" endpoint status -w table
