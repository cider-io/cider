#!/bin/sh

echo "Building cider for linux"
(cd ..; go clean; GOOS=linux GOARCH=amd64 go build)


for i in {01..10}; do
    host=sp21-cs525-g17-$(printf "%02d" $i).cs.illinois.edu
    echo "Transfer test script to $host"
    ssh $host 'rm *.log'
    ssh $host 'rm cider_test.py'
    scp cider_test.py $host:.

    echo "Kill cider if already running on $host"
    ssh $host 'kill $(pgrep cider)'

    echo "Transfer cider to $host"
    ssh $host 'rm cider'
    scp ../cider $host:.
    
    echo "Start cider on $host"
    ssh $host "./cider >/dev/null 2>&1 &"
    
    echo "Verify that cider is running on $host"
    curl -X GET -i http://$host:6143/tasks
    echo ""
done

rnum=$((( RANDOM % 10 )  + 1 ))
random_host=sp21-cs525-g17-$(printf "%02d" $rnum).cs.illinois.edu

echo $random_host
for i in {01..10}; do
    host=sp21-cs525-g17-$(printf "%02d" $i).cs.illinois.edu
    if [ $host != $random_host ]
    then
        ssh $host "./cider_test.py $random_host >/dev/null 2>&1 &"
    fi
done
