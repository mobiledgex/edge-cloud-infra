#!/bin/bash

MAIN_ANSIBLE_VAULT_PREFIX='ansible-mex-vault'
PERSONAL_ANSIBLE_VAULT='personal-ansible-vault.yml'
DEFAULT_PLAYBOOK='mexplat.yml'

USAGE="usage: $0 [options] <environment> [<target>]

  -c	confirm before running playbook
  -l	list available targets
  -n	dry-run mode
  -p	playbook (default: \"$DEFAULT_PLAYBOOK\")
  -t	tags (comma-separated)
  -y	skip confirmation prompts

  -h	display this help message

example: $0 -n staging console"

# See: https://github.com/ansible/ansible/issues/49207
export OBJC_DISABLE_INITIALIZE_FORK_SAFETY=YES

DRYRUN=false
LIST=false
CONFIRM=false
ASSUME_YES=false
PLAYBOOK_FORCED=
TAGS=
while getopts ':cChlnp:t:y' OPT; do
	case "$OPT" in
	c)	CONFIRM=true ;;
	n|C)	DRYRUN=true ;;
	l)	LIST=true ;;
	p)	PLAYBOOK_FORCED="$OPTARG" ;;
	t)	TAGS="$OPTARG" ;;
	y)	ASSUME_YES=true ;;
	h)	echo "$USAGE"
		exit 0
		;;
	*)	echo "unknown option or missing argument: $OPTARG" >&2
		echo "$USAGE" >&2
		exit 1
		;;
	esac
done
shift $(( OPTIND - 1 ))

ENVIRON="$1"; shift
TARGET="$1"; shift

if [[ -z "$ENVIRON" ]]; then
	echo "$USAGE" >&2
	exit 1
fi

if [[ ! -e "$ENVIRON" ]]; then
	echo "$ENVIRON: inventory not found" >&2
	exit 2
fi

if [[ -n "$PLAYBOOK_FORCED" && ! -f "$PLAYBOOK_FORCED" ]]; then
	echo "$PLAYBOOK_FORCED: playbook not found" >&2
	exit 2
fi

[[ "$ENVIRON" == main ]] && CONFIRM=true

# List mode
"$LIST" && exec ansible-inventory -i "$ENVIRON" --graph

ARGS=()
$DRYRUN && ARGS+=( '--check' )

MAIN_VAULT="${MAIN_ANSIBLE_VAULT_PREFIX}-${ENVIRON}.yml"
[[ ! -f "$MAIN_VAULT" ]] && MAIN_VAULT="${MAIN_ANSIBLE_VAULT_PREFIX}.yml"
[[ -f "$MAIN_VAULT" ]] && ARGS+=( -e "@${MAIN_VAULT}" )

# Add personal ansible vault to command line, if present
[[ -f "$PERSONAL_ANSIBLE_VAULT" ]] && ARGS+=( -e "@${PERSONAL_ANSIBLE_VAULT}" )

# Limit to specified target
[[ -n "$TARGET" ]] && ARGS+=( -l "$TARGET" )

# Tags
[[ -n "$TAGS" ]] && ARGS+=( -t "$TAGS" )

# Inventory
ARGS+=( -i "$ENVIRON" )

# Playbook
if [[ -n "$PLAYBOOK_FORCED" ]]; then
	PLAYBOOK="$PLAYBOOK_FORCED"
else
	PLAYBOOK="${TARGET}.yml"
	[[ -f "$PLAYBOOK" ]] || PLAYBOOK="$DEFAULT_PLAYBOOK"
fi
ARGS+=( "$PLAYBOOK" )

echo
$DRYRUN && echo -n " [DRYRUN] "
echo "=> ansible-playbook ${ARGS[*]}"
echo
if $CONFIRM && ! $DRYRUN && ! $ASSUME_YES; then
	read -p "Are you sure you want to run this command in the \"[1;31m$ENVIRON[0m\" environment? (n) " RESP
	case "$RESP" in
	y*|Y*)	echo ;;
	*)	echo "Aborting..."; exit 0 ;;
	esac
fi

ansible-playbook "${ARGS[@]}"; RC=$?

if $DRYRUN; then
	echo
	echo "=> REMEMBER: This was a dry run!"
	echo
fi

exit "$RC"
