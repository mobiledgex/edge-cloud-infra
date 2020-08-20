#!/bin/bash
ARTIFACTORY_BASEURL='https://artifactory.mobiledgex.net'
ARTIFACTORY_USER='packer'
ARTIFACTORY_ARTIFACTS_TAG='2020-04-27'
CLOUD_IMAGE='ubuntu-18.04-server-cloudimg-amd64.img'
OUTPUT_IMAGE_NAME='mobiledgex'

: ${CLOUD_IMAGE_TAG:=ubuntu-18.04-server-cloudimg-amd64}
: ${VAULT:=vault-main.mobiledgex.net}
: ${FLAVOR:=m4.small}
: ${FORCE:=no}
: ${TRACE:=no}
: ${DEBUG:=false}

GITTAG=$( git describe --tags )
[[ -z "$TAG" ]] && TAG="$GITTAG"

USAGE="usage: $( basename $0 ) <options>

 -d               Run in debug mode
 -f <flavor>      Image flavor (default: \"$FLAVOR\")
 -i <image-tag>   Glance source image tag (default: \"$CLOUD_IMAGE_TAG\")
 -o <output-tag>  Output image tag (default: same as tag below)
 -p <platform>    Output platform flavor; one of \"openstack\" (default) or \"vsphere\"
 -t <tag>         Image tag name (default: \"$TAG\")
 -F               Ignore source image checksum mismatch
 -T               Print trace debug messages during build
 -u <artf-user>   Build as this Artifactory user (default: \"$ARTIFACTORY_USER\")

 -h               Display this help message
"

while getopts ":dhf:i:o:p:t:FTu:" OPT; do
	case "$OPT" in
	d) DEBUG=true ;;
	h) echo "$USAGE"; exit 0 ;;
	i) CLOUD_IMAGE_TAG="$OPTARG" ;;
	f) FLAVOR="$OPTARG" ;;
	o) OUTPUT_TAG="$OPTARG" ;;
	p) OUTPUT_PLATFORM="$OPTARG" ;;
	t) TAG="$OPTARG" ;;
	F) FORCE=yes ;;
	T) TRACE=yes ;;
	u) ARTIFACTORY_USER="$OPTARG" ;;
	esac
done
shift $(( OPTIND - 1 ))

die() {
	echo "ERROR: $*" >&2
	exit 2
}

[[ -z "$OUTPUT_PLATFORM" ]] && OUTPUT_PLATFORM=openstack
case "$OUTPUT_PLATFORM" in
	openstack)	true ;;
	vsphere)	TAG="${TAG%-vsphere}-vsphere" ;;
	*)		die "Unknown platform type: $OUTPUT_PLATFORM" ;;
esac

TAG=${TAG#v}
[[ -z "$OUTPUT_TAG" ]] && OUTPUT_TAG="v$TAG"

ARTIFACTORY_APIKEY_FILE="${HOME}/.mobiledgex/artifactory.apikey"
if [[ -f "$ARTIFACTORY_APIKEY_FILE" ]]; then
	ARTIFACTORY_APIKEY=$( head -n 1 "$ARTIFACTORY_APIKEY_FILE" )
else
	read -s -p "Artifactory password/api-key: " ARTIFACTORY_APIKEY
	echo
fi

VAULT_PATH="secret/accounts/baseimage"
export VAULT_ADDR="https://${VAULT}"
if ! vault token lookup >/dev/null 2>&1; then
	echo "Logging in to $VAULT_ADDR"
	vault login -method=github
	[[ $? -eq 0 ]] || die "Failed to log in to vault: $VAULT_ADDR"
	echo
fi

ROOT_PASS=$( vault kv get -field=value "${VAULT_PATH}/password" )
GRUB_PW_HASH=$( vault kv get -field=grub_pw_hash "${VAULT_PATH}/password" )
TOTP_KEY=$( vault kv get -field=value "${VAULT_PATH}/totp-key" )
if [[ -z "$ROOT_PASS" || -z "$GRUB_PW_HASH" || -z "$TOTP_KEY" ]]; then
	die "Unable to read vault secrets: ${VAULT} ${VAULT_PATH}"
fi

jq_VERSION=$( jq --version 2>/dev/null )
openstack_VERSION=$( openstack --version 2>&1 | grep '^openstack' | awk '{print $2}' )
packer_VERSION=$( packer --version 2>/dev/null )

echo "TOOL VERSIONS:"; echo
for TOOL in jq openstack packer; do
	VERS=$( eval echo \$${TOOL}_VERSION )
	[[ -z "$VERS" ]] && die "Required tool not found: $TOOL"
	echo "  ${TOOL}: ${VERS}"
done
echo

OUTPUT_IMAGE_NAME="${OUTPUT_IMAGE_NAME}-${OUTPUT_TAG}"

ARTIFACTORY_CLOUD_IMAGE_PATH="baseimage-build/${ARTIFACTORY_ARTIFACTS_TAG}/${CLOUD_IMAGE}"
ARTIFACTORY_CLOUD_IMAGE_CHECKSUM=$( curl -sSL -u "${ARTIFACTORY_USER}:${ARTIFACTORY_APIKEY}" \
	"${ARTIFACTORY_BASEURL}/api/storage/${ARTIFACTORY_CLOUD_IMAGE_PATH}" \
	| jq -er '.checksums.md5' )
[[ $? -ne 0 ]] && die "Failed to retrieve cloud image checksum from artifactory: TAG=$ARTIFACTORY_ARTIFACTS_TAG"

SRC_IMG=$( openstack image list -c ID -c Name -f value \
	| grep " ${CLOUD_IMAGE_TAG}$" \
	| cut -d' ' -f1 )
if [[ -z "$SRC_IMG" ]]; then
	openstack image list >/dev/null 2>&1
	[[ $? -ne 0 ]] \
		&& die "Unable to locate openstack source image; openrc not sourced?"
	die "Unable to locate source image in glance: $CLOUD_IMAGE_TAG"
fi

SRC_IMG_CHECKSUM=$( openstack image show -c checksum -f value "$SRC_IMG" )

if [[ "$SRC_IMG_CHECKSUM" != "$ARTIFACTORY_CLOUD_IMAGE_CHECKSUM" ]]; then
	if [[ "$FORCE" == yes ]]; then
		echo "Ignoring checksum mismatch of image in glance: $CLOUD_IMAGE_TAG" >&2
	else
		die "Cloud image checksum does not match image in glance: $CLOUD_IMAGE_TAG"
	fi
fi

NETWORK=$( openstack network list -c ID -c Name -f value \
	| grep ' external-network-shared$' \
	| cut -d' ' -f1 )
[[ -z "$NETWORK" ]] && die "Unable to locate openstack network details"

cat <<EOT
BUILD PARAMETERS:

  Source Image UUID: $SRC_IMG ("$CLOUD_IMAGE_TAG")
       Network UUID: $NETWORK
     New Image Name: $OUTPUT_IMAGE_NAME
             Flavor: $FLAVOR
   Artifactory User: $ARTIFACTORY_USER
    Output Platform: $OUTPUT_PLATFORM

EOT

read -p "Build? (yN) " RESP
case "$RESP" in
	y*|Y*)	true ;;
	*)	echo "Aborting build..."; exit 1 ;;
esac

CMDLINE=( packer build -on-error=ask )
$DEBUG && CMDLINE+=( -debug )
PACKER_LOG=1 "${CMDLINE[@]}" \
	-var "OUTPUT_IMAGE_NAME=$OUTPUT_IMAGE_NAME" \
	-var "SRC_IMG=$SRC_IMG" \
	-var "SRC_IMG_CHECKSUM=$SRC_IMG_CHECKSUM" \
	-var "NETWORK=$NETWORK" \
	-var "ARTIFACTORY_USER=$ARTIFACTORY_USER" \
	-var "ARTIFACTORY_APIKEY=$ARTIFACTORY_APIKEY" \
	-var "ARTIFACTORY_ARTIFACTS_TAG=$ARTIFACTORY_ARTIFACTS_TAG" \
	-var "ROOT_PASS=$ROOT_PASS" \
	-var "GRUB_PW_HASH=$GRUB_PW_HASH" \
	-var "TOTP_KEY=$TOTP_KEY" \
	-var "TAG=$TAG" \
	-var "GITTAG=$GITTAG" \
	-var "FLAVOR=$FLAVOR" \
	-var "TRACE=$TRACE" \
	-var "MEX_BUILD=$( git describe --long --tags )" \
	-var "OUTPUT_PLATFORM=$OUTPUT_PLATFORM" \
	packer_template.mobiledgex.json

if [[ $? -ne 0 ]]; then
	echo "Failed to build base image!" >&2
	exit 2
fi

echo
read -p "Upload to Artifactory? (yN) " RESP
case "$RESP" in
	y*|Y*)	true ;;
	*)	echo "NOT uploading to Artifactory"; exit 0 ;;
esac

IMAGE_FNAME="${OUTPUT_IMAGE_NAME}.qcow2"

echo
echo "Downloading image from glance: $IMAGE_FNAME ..."
openstack image save --file "$IMAGE_FNAME" "$OUTPUT_IMAGE_NAME"

GLANCE_CHECKSUM=$( openstack image show -c checksum -f value "$OUTPUT_IMAGE_NAME" )
LOCAL_CHECKSUM=$( openssl md5 "$IMAGE_FNAME" | awk '{print $NF}' )
if [[ "$LOCAL_CHECKSUM" != "$GLANCE_CHECKSUM" ]]; then
	echo "Error downloading \"$OUTPUT_IMAGE_NAME\" image; checksum mismatch" >&2
	exit 2
fi

echo
echo "Uploading image to Artifactory..."
LOCAL_SHA1SUM=$( openssl sha1 "$IMAGE_FNAME" | awk '{print $NF}' )
curl -sSL -XPUT -u "${ARTIFACTORY_USER}:${ARTIFACTORY_APIKEY}" \
	-H "X-Checksum-MD5:$LOCAL_CHECKSUM" \
	-H "X-Checksum-Sha1:$LOCAL_SHA1SUM" \
	-T "$IMAGE_FNAME" \
	"${ARTIFACTORY_BASEURL}/baseimages/${IMAGE_FNAME}"
if [[ $? -ne 0 ]]; then
	echo "Error uploading image to Artifactory" >&2
	exit 2
fi

rm "$IMAGE_FNAME"
