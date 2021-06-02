#!/bin/bash

ARTIFACTORY_BASE="https://artifactory.mobiledgex.net/artifactory"

die() {
	echo "ERROR: $*" >&2
	exit 2
}

log() {
	echo "============================================================"
	echo "$*"
	echo
}

artf_call() {
	_call=$1; shift
	curl -sf -H "X-JFrog-Art-Api:$ARTIFACTORY_APIKEY" \
		"${ARTIFACTORY_BASE}/$_call" "$@"
}

### Main ###

[[ -z "$BASE_IMAGE_NAME" ]] && die "BASE_IMAGE_NAME not set"
[[ -z "$OPENSTACK_INSTANCE" ]] && die "OPENSTACK_INSTANCE not set"
[[ -z "$ARTIFACTORY_APIKEY" ]] && die "ARTIFACTORY_APIKEY not set"

COMPRESSED_BASE_IMAGE_NAME="${BASE_IMAGE_NAME%_uncompressed}"

OPENRC="$HOME/.cloudlets/${OPENSTACK_INSTANCE}_cloudlet/openrc.mex"
[[ -f "$OPENRC" ]] || die "OpenRC not found: $OPENRC"

. ${OPENRC}
[[ -z "$OS_AUTH_URL" ]] && die "Failed to source OpenRC"

CHECKSUM=$( openstack image show -c checksum -f value "$BASE_IMAGE_NAME" )
[[ -z "$CHECKSUM" ]] && die "Failed to get image checksum: $BASE_IMAGE_NAME"

ARTF_VERIFY=$( artf_call "api/storage/baseimages/README.txt" | jq -r .path )
[[ "$ARTF_VERIFY" == /README.txt ]] || die "Failed to access Artifactory"

ARTIFACT_PATH="baseimages/${COMPRESSED_BASE_IMAGE_NAME}.qcow2"
ARTF_VERIFY=$( artf_call "api/storage/${ARTIFACT_PATH}" | jq -r .uri )
[[ -n "$ARTF_VERIFY" ]] && die "Base image exists: $ARTF_VERIFY"

IMAGE="$PWD/image.qcow2"
COMPRESSED_IMAGE="$PWD/image-compressed.qcow2"
trap 'rm -f "$IMAGE" "$COMPRESSED_IMAGE"' EXIT

log "Downloading image $BASE_IMAGE_NAME"
openstack image save --file "$IMAGE" "$BASE_IMAGE_NAME"
[[ $? -eq 0 ]] || die "Failed to download image: $BASE_IMAGE_NAME"

log "Validating checksum"
IMAGE_CHECKSUM=$( md5sum "$IMAGE" | awk '{print $1}' )
[[ "$IMAGE_CHECKSUM" != "$CHECKSUM" ]] && die "Checksum mismatch: $BASE_IMAGE_NAME"

if [[ "$BASE_IMAGE_NAME" == "$COMPRESSED_BASE_IMAGE_NAME" ]]; then
	log "Not compressing image"
	echo "Assuming image is already compressed as it does not have the _\"uncompressed\" suffix"
	mv "$IMAGE" "$COMPRESSED_IMAGE"
else
	log "Compressing image"
	qemu-img convert -c -O qcow2 "$IMAGE" "$COMPRESSED_IMAGE"
fi

log "Uploading image"
COMPRESSED_IMAGE_CHECKSUM=$( md5sum "$COMPRESSED_IMAGE" | awk '{print $1}' )
ARTF_CHECKSUM=$( artf_call "${ARTIFACT_PATH}" -T "$COMPRESSED_IMAGE" | jq -r .checksums.md5 )
[[ -z "$ARTF_CHECKSUM" ]] && die "Error uploading image: $ARTIFACT_PATH"
[[ "$ARTF_CHECKSUM" != "$COMPRESSED_IMAGE_CHECKSUM" ]] \
	&& die "Upload error; checksum mismatch: $ARTIFACT_PATH"

log "Image uploaded to Artifactory: $ARTIFACT_PATH"
log "You can remove the uncompressed base image from Glance now: $BASE_IMAGE_NAME"
