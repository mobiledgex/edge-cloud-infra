#!/bin/bash
# must be run as root

USAGE="usage: $( basename $0 ) <options>

 -s <chef-server-url> Chef Server URL
 -n <node-name>        Chef client node-name

 -h                    Display this help message
"
while getopts ":hs:n:" OPT; do
        case "$OPT" in
        h) echo "$USAGE"; exit 0 ;;
        s) CHEFSERVERURL="$OPTARG" ;;
        n) NODENAME="$OPTARG" ;;
        esac
done
shift $(( OPTIND - 1 ))

die() {
        echo "ERROR: $*" >&2
        exit 2
}

[[ -z $CHEFSERVERURL ]] && die "Missing chef server URL"
[[ -z $NODENAME ]] && die "Missing node name"

cat > /etc/chef/client.rb <<EOT
log_level              :info
ssl_verify_mode        :verify_none
client_key             "/home/ubuntu/client.pem"
chef_server_url        "$CHEFSERVERURL"
node_name              "$NODENAME"
Chef::Log::Formatter.show_time = true
EOT

systemctl restart chef-client
[[ $? -ne 0 ]] && die "Failed to restart chef-client service"
