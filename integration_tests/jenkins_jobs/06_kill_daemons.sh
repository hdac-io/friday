#!/bin/bash

kill -9 $(cat $HOME/node.pid)
kill -9 $(cat $HOME/ee.pid)
