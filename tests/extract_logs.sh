#!/bin/sh

for i in {01..10}; do
    host=sp21-cs525-g17-$(printf "%02d" $i).cs.illinois.edu
    echo "Extracting logs from $host"
    mkdir $host
    ssh $host "grep 'METRIC' cider.log > metrics.log"
    scp $host:*.log $host/.
done
