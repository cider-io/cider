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


def single_request(host):
    size = 1000000

    ran_floats = [random.uniform(-10, 80) for _ in range(size)]

    base_url = f'http://{host}:6143'

    data = {}
    data["data"] = ran_floats
    data["function"] = "sum"

    logger.info(f'Sending data to {host}.')
    response = requests.put(base_url+'/tasks', json=data)

    try:
        response.raise_for_status()
    except HTTPError as e:
        logger.error(f'HTTP Error {e}')

    response_dict = response.json()["results"]
    assert response_dict['function'] == 'sum', 'Invalid function'
    id = response_dict['id']

    logger.info(
        f'Getting task status from {host} for id {id}, wait until `Succeeded` as status')

    loops = 1
    while True:
        response = requests.get(base_url + f'/tasks/{id}/status')
        try:
            response.raise_for_status()
        except HTTPError as e:
            logger.error(f'HTTP Error {e}')
        status = response.json()["results"]
        if status == "Stopped":
            logger.info(
                f'{host} finished task with id {id} in {loops} loop(s).')
            break
        loops += 1

    logger.info(f'Getting result from {host} for id {id}')
    response = requests.get(base_url + f'/tasks/{id}/result')
    try:
        response.raise_for_status()
    except HTTPError as e:
        logger.error(f'HTTP Error {e}')
    results = response.json()["results"]
    result = results['result']
    error = results['error']
    logger.info(f'Result: {result}, Error: {error}')

    expected = sum(ran_floats)
    logger.info(f'Expected: {expected}')

    assert result == expected, 'Incorrect result'


if __name__ == '__main__':
    host = sys.argv[1]
    logger.info(f"Targeting remote host: {host}")
    for i in range(100):
        logger.info(f'Iteration {i}')
        single_request(host)
