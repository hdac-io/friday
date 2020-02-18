#!/bin/bash

# No predefined path /usr/sbin at CentOS
export PATH=/usr/sbin:$PATH

export BACKUP_BUILD_ID=$BUILD_ID
echo $BACKUP_BUILD_ID

daemonize -E BUILD_ID=dontKillMe -e /var/log/friday/ee.$BACKUP_BUILD_ID.err -o /var/log/friday/ee.$BACKUP_BUILD_ID.log -p /tmp/ee.pid $HOME/friday/CasperLabs/execution-engine/target/release/casperlabs-engine-grpc-server $HOME/.casperlabs/.casper-node.sock
daemonize -E BUILD_ID=dontKillMe -e /var/log/friday/node.$BACKUP_BUILD_ID.err -o /var/log/friday/node.$BACKUP_BUILD_ID.log -p /tmp/node.pid $HOME/go/bin/nodef start
