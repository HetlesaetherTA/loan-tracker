#!/bin/sh

./tmp/loan-tracker &

sleep 0.5

./tmp/web-server &

wait
