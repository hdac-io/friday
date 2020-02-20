#!/bin/bash

cd $HOME/friday/integration_tests && \
pip3 install -r ./requirements.txt && \
PATH=$PATH:$HOME/go/bin && \
python3 multinode_test_setup.py
