#!/bin/bash

cd $HOME/friday/integration_tests && \
pip3 install -r ./requirements.txt && \
PATH=$PATH:$HOME/go/bin && \
python3 config_setting.py

scp -i ~/ci_nodes.pem ~/.nodef/config/genesis.json opc@132.145.83.49:~/.nodef/config
scp -i ~/ci_nodes.pem -r ~/.nodef/contracts opc@132.145.83.49:~/.nodef
scp -i ~/ci_nodes.pem ~/config.toml opc@132.145.83.49:~/.nodef/config

scp -i ~/ci_nodes.pem ~/.nodef/config/genesis.json opc@150.136.172.164:~/.nodef/config
scp -i ~/ci_nodes.pem -r ~/.nodef/contracts opc@150.136.172.164:~/.nodef
scp -i ~/ci_nodes.pem ~/config.toml opc@150.136.172.164:~/.nodef/config

scp -i ~/ci_nodes.pem ~/.nodef/config/genesis.json ubuntu@140.238.73.77:~/.nodef/config
scp -i ~/ci_nodes.pem -r ~/.nodef/contracts ubuntu@140.238.73.77:~/.nodef
scp -i ~/ci_nodes.pem ~/config.toml ubuntu@140.238.73.77:~/.nodef/config