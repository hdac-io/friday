#!/bin/bash

kill -9 $(cat /tmp/node.pid)
kill -9 $(cat /tmp/ee.pid)
