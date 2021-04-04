#!/usr/bin/python3
import requests
from requests.exceptions import HTTPError
import random
import json
import time
import sys
import logging

logging.basicConfig(filename="test.log",
                    format='%(asctime)s %(message)s',
                    filemode='w')

logger = logging.getLogger()

logger.setLevel(logging.INFO)

DATA_SIZE = 1000000
ITERATIONS = 100

hosts = [
    'sp21-cs525-g17-{:02}.cs.illinois.edu'.format(i + 1) for i in range(10)]


def single_request(host):
    ran_floats = [random.uniform(-10, 80) for _ in range(DATA_SIZE)]

    base_url = f'http://{host}:6143'

    data = {}
    data["data"] = ran_floats
    data["function"] = "sum"

    logger.info(f'Sending data to {host}.')
    response = requests.put(base_url+'/tasks', json=data)

    # try:
    #     response.raise_for_status()
    # except HTTPError as e:
    #     logger.error(f'HTTP Error {e}')

    response_dict = response.json()
    # assert response_dict['function'] == 'sum', 'Invalid function'
    id = response_dict['id']

    logger.info(
        f'Getting task status from {host} for id {id}, wait until `Succeeded` as status')

    loops = 1
    while True:
        response = requests.get(base_url + f'/tasks/{id}')
        # try:
        #     response.raise_for_status()
        # except HTTPError as e:
        #     logger.error(f'HTTP Error {e}')
        response_dict = response.json()
        if response_dict["status"] == "Stopped":
            logger.info(
                f'{host} finished task with id {id} in {loops} loop(s).')
            break
        loops += 1

    logger.info(f'Getting result from {host} for id {id}')
    response = requests.get(base_url + f'/tasks/{id}/result')
    # try:
    #     response.raise_for_status()
    # except HTTPError as e:
    #     logger.error(f'HTTP Error {e}')
    results = response.json()
    result = results['result']
    error = results['error']
    logger.info(f'Result: {result}, Error: {error}')

    expected = sum(ran_floats)
    logger.info(f'Expected: {expected}')

    assert result == expected, 'Incorrect result'


def test_pick_same_one_random(rseeds):
    # All the VMs will pick the same random node
    # to send the data for compute as we use
    # the common seed
    random.seed(rseeds[0])
    host = random.sample(hosts, 1)[0]

    # Use the unique seed for data randomness
    # across VMs
    random.seed(rseeds[1])
    time.sleep(random.random()*10)

    for i in range(ITERATIONS):
        logger.info(f'Iteration {i+1}')
        single_request(host)


def test_pick_same_three_random(rseeds):
    # All the VMs will pick the same three
    # random nodes to send the data for
    # compute as we use the common seed
    random.seed(rseeds[0])
    target_hosts = random.sample(hosts, 3)

    # Use the unique seed for data randomness
    # across VMs
    random.seed(rseeds[1])
    time.sleep(random.random()*10)

    for i in range(ITERATIONS):
        host = random.sample(target_hosts, 1)[0]
        logger.info(f'Iteration {i+1}')
        single_request(host)


def test_pick_same_five_random(rseeds):
    # All the VMs will pick the same five
    # random nodes to send the data for
    # compute as we use the common seed
    random.seed(rseeds[0])
    target_hosts = random.sample(hosts, 5)

    # Use the unique seed for data randomness
    # across VMs
    random.seed(rseeds[1])
    time.sleep(random.random()*10)

    for i in range(ITERATIONS):
        host = random.sample(target_hosts, 1)[0]
        logger.info(f'Iteration {i+1}')
        single_request(host)


def test_pick_single_random(rseeds):
    # All the VMs will pick a single
    # not necessarily same random node
    # to send the data for compute
    # as we use the unique seed only
    random.seed(rseeds[1])
    time.sleep(random.random()*10)
    host = random.sample(hosts, 1)[0]

    for i in range(ITERATIONS):
        logger.info(f'Iteration {i+1}')
        single_request(host)


def test_pick_one_random_per_iteration(rseeds):
    # All the VMs will pick one random
    # node, not necessarily the same one,
    # each iteration to send the data for
    # compute as we use the unique seed only
    random.seed(rseeds[1])
    time.sleep(random.random()*10)

    for i in range(ITERATIONS):
        host = random.sample(hosts, 1)[0]
        logger.info(f'Iteration {i+1}')
        single_request(host)


test_map = {
    1: test_pick_same_one_random,
    2: test_pick_same_three_random,
    3: test_pick_same_five_random,
    4: test_pick_single_random,
    5: test_pick_one_random_per_iteration,
}

if __name__ == '__main__':
    test = int(sys.argv[1])
    # Random seeds to be used for various tests
    # rseed1 is common for all VMs, rest can be different
    rseeds = [int(a) for a in sys.argv[2:]]
    test_map[test](rseeds)
