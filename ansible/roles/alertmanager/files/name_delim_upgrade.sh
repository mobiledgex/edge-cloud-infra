#!/bin/bash
cd /var/tmp/alertmanager
# replace deleimeter in receiver name: emailClusterAlertReceiver-mexadmin-error-email -> emailClusterAlertReceiver::mexadmin::error::email
cat config.yml | sed -E 's/(- name:.+|- receiver:.+)-(.+)-(.+)-(.+)/\1::\2::\3::\4/g' > config.yml.v2
cp config.yml config.yml.old
cp config.yml.v2 config.yml
