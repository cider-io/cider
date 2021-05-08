#!/usr/bin/python3
import os
import datetime
import json

log_files = {}
for subdir, dirs, files in os.walk('.'):
    if subdir != '.':
        host = subdir.split(os.sep)[1][15:17]
        log_files[host] = {}
        for file in files:
            if file.endswith('log'):
                with open(os.path.join(subdir, file)) as f:
                    log_files[host][file.split('.')[0]] = f.readlines()

test_metrics = {}

for host in log_files:
    if 'metrics' in log_files[host].keys():
        for metric in log_files[host]['metrics']:
            try:
                metric = json.loads(metric.split()[-1])
                ds = datetime.datetime.strptime(metric["start"], '%H:%M:%S.%f')
                de = datetime.datetime.strptime(metric["end"], '%H:%M:%S.%f')
                test_metrics[metric["id"]] = {}
                test_metrics[metric["id"]]["compute_time"] = (
                    de - ds).total_seconds()
                test_metrics[metric["id"]]['compute_node'] = host
                test_metrics[metric["id"]]['start_time'] = metric["start"]
                test_metrics[metric["id"]]['end_time'] = metric["end"]
            except Exception as e:
                boom = True
                print(e)


with open('test.csv', 'w') as fw:
    line = 'time,compute_node\n'
    fw.write(line)


with open('test.csv', 'a') as fa:
    for k in test_metrics.keys():
        ds = datetime.datetime.strptime(
            test_metrics[k]['start_time'], '%H:%M:%S.%f')
        de = datetime.datetime.strptime(
            test_metrics[k]['end_time'], '%H:%M:%S.%f')
        while ds < de:
            line = (
                f"{ds.strftime('%H:%M:%S')},{test_metrics[k]['compute_node']}\n")
            fa.write(line)
            ds += datetime.timedelta(seconds=1)
