#!/bin/sh

echo "Building cider for linux"
(cd ..; go clean; GOOS=linux GOARCH=amd64 go build)

# Kill cider already running
./stop_test.sh

for i in {01..10}; do
    host=sp21-cs525-g17-$(printf "%02d" $i).cs.illinois.edu
    echo "Transfer test script to $host"
    ssh $host 'rm *.log'
    ssh $host 'rm cider_test.py'
    scp cider_test.py $host:.

    echo "Transfer cider to $host"
    ssh $host 'rm cider'
    scp ../cider $host:.
    
    echo "Start cider on $host"
    if [ $i -le 7 ]
    then
        echo " Resource constrained node"
        ssh $host "./cider  --resource-constrained >/dev/null 2>&1 &"
    else
        echo " Regular node"
        ssh $host "./cider >/dev/null 2>&1 &"
    fi
    
    echo "Verify that cider is running on $host"
    ssh $host 'curl -X GET -i http://localhost:6143/tasks'
    echo ""
done

# test_map = {
#     1: test_pick_same_one_random,
#     2: test_pick_same_three_random,
#     3: test_pick_same_five_random,
#     4: test_pick_single_random,
#     5: test_pick_one_random_per_iteration,
#     6: test_pick_one_random_per_iteration_with_waits,
#     7: test_end_to_end,
# }
echo "Using test option $1"

rnum1=$((( RANDOM % 100 )  + 1 ))
for i in {01..10}; do
    echo "Starting test on VM $(printf "%02d" $i)"
    # Generate additional random seed for use in test
    rnum2=$((( RANDOM % 100 )  + 1 ))
    host=sp21-cs525-g17-$(printf "%02d" $i).cs.illinois.edu
    ssh $host "./cider_test.py $1 $rnum1 $rnum2 >error.log 2>&1 &"
done
