#!/usr/bin/python3
import http.client
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

    conn = http.client.HTTPConnection(host, 6143, timeout=3000)

    id = hash(str(ran_floats))
    data = {}
    data["id"] = "{}".format(id)
    data["data"] = ran_floats
    data["function"] = "sum"

    data_to_send = json.dumps(data)

    logger.info(f'Sending data to {host} for id {id}')
    conn.request("PUT", "/tasks", data_to_send)
    response = conn.getresponse()
    assert response.status == 200, 'Invalid status'
    _ = response.read()

    logger.info(f'Getting task status from {host} for id {id}')
    conn.request("GET", "/tasks/{}".format(id))
    response = conn.getresponse()
    result = json.loads(response.read())
    status = result["results"]["status"]
    assert status == "Succeeded", 'mismatch: {}, itr {}'.format(status, iter)

    logger.info(f'Getting result from {host} for id {id}')
    conn.request("GET", "/tasks/{}/result".format(id))
    response = conn.getresponse()

    assert response.status == 200, 'Invalid status'
    result = json.loads(response.read())

    print(result)

    expected = sum(ran_floats)
    logger.info(f'Expected: {expected}')

    assert result["results"] == expected, 'Incorrect result'

    conn.close()

    logger.info('Result: {}'.format(result["results"]))


if __name__ == '__main__':
    host = sys.argv[1]
    logger.info(f"Targeting remote host: {host}")
    for i in range(100):
        logger.info(f'Iteration {i}')
        single_request(host)
