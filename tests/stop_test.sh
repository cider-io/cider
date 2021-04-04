#!/bin/sh


for i in {01..10}; do
    host=sp21-cs525-g17-$(printf "%02d" $i).cs.illinois.edu
    echo "Kill cider if already running on $host"
    ssh $host 'kill $(pgrep cider)'
done