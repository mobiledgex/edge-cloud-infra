#!/bin/bash

VERSION=$1
OUT=package-version.go

cat <<EOF > $OUT.tmp
package version

var MobiledgeXPackageVersion = "$VERSION"
EOF

gofmt $OUT.tmp > $OUT
rm $OUT.tmp
