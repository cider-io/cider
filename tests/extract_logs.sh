#!/bin/sh

for i in {07..10}; do
    host=sp21-cs525-g17-$(printf "%02d" $i).cs.illinois.edu
    echo "Extracting logs from $host"
    mkdir $host
    scp $host:*.log $host/.
done
