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

for host in log_files.keys():
    try:
        if 'test' in log_files[host].keys():
            for i in range(len(log_files[host]['test'])):
                if 'Iteration' in log_files[host]['test'][i]:
                    it = log_files[host]['test'][i].split()[-1]
                    send_line = log_files[host]['test'][i+1]
                    id = log_files[host]['test'][i+4].split()[-1]
                    result_line = log_files[host]['test'][i+5]
                    _, send_time, _, _, _, compute_node = send_line.split()
                    result_time = result_line.split()[1]
                    ds = datetime.datetime.strptime(send_time, '%H:%M:%S,%f')
                    dr = datetime.datetime.strptime(result_time, '%H:%M:%S,%f')
                    total_duration = (dr - ds).total_seconds()
                    test_metrics[id] = {
                        'iteration': it,
                        'request_node': host,
                        'compute_node': compute_node[15:17],
                        'total_duration': total_duration
                    }
                    i += 6
    except Exception as e:
        print(e)

for host in log_files:
    if 'metrics' in log_files[host].keys():
        for metric in log_files[host]['metrics']:
            try:
                metric = json.loads(metric.split()[-1])
                ds = datetime.datetime.strptime(metric["start"], '%H:%M:%S.%f')
                de = datetime.datetime.strptime(metric["end"], '%H:%M:%S.%f')
                test_metrics[metric["id"]]["compute_time"] = (
                    de - ds).total_seconds()
                test_metrics[metric["id"]]["function"] = metric["function"]
                test_metrics[metric["id"]]["datasize"] = metric["datasize"]
            except Exception as e:
                boom = True
                print(e)


with open('test.csv', 'w') as fw:
    line = 'iteration,request_node,compute_node,function,datasize,compute_time,total_duration\n'
    fw.write(line)


with open('test.csv', 'a') as fa:
    for k in test_metrics.keys():
        line = (f"{test_metrics[k]['iteration']},{test_metrics[k]['request_node']},"
                f"{test_metrics[k]['compute_node']},{test_metrics[k]['function']},"
                f"{test_metrics[k]['datasize']},{test_metrics[k]['compute_time']},"
                f"{test_metrics[k]['total_duration']}\n")
        fa.write(line)
