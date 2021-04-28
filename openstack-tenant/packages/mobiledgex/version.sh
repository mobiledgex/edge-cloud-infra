#!/bin/bash

VERSION=$1
OUT=package-version.go

OUT_TMP=$( mktemp )
trap 'rm -f "$OUT_TMP"' EXIT

cat <<EOF > $OUT_TMP
package version
var MobiledgeXPackageVersion = "$VERSION"
EOF

gofmt $OUT_TMP > $OUT
