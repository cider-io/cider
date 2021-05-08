# CIDER: **C**ooperative **I**n-home **D**istributed **E**fficient **R**esource Allocation Protocol

## Quickstart
```bash
go build .

# To simulate a resource-constrained IoT device, that will not run tasks:
./cider --introducer {Introducer IP or hostname} --resource-constrained

# To simulate a capable compute device:
./cider --introducer {Introducer IP or hostname}
```

## API User Guide
```bash
# Get all tasks running on a node
curl -i {Node IP or hostname}:6143/tasks

# Deploy a task to a node
# Note: This may return a 200 or a 303 (Temporary Redirect)
# If 303: Follow the provided URL in the Location header
# If 200: use the returned {Task ID} in future requests
curl -iX PUT {Node IP or hostname}:6143/tasks --data '{"function": "sum", "data": [1,2,3,4,5]}'

# Fetch the status of a task
curl -i {Node IP or hostname}:6143/tasks/{Task ID}

# Send control signals (e.g. abort) to a Running task
# Note: Only abort is supported at the moment
curl -iX PUT {Node IP or hostname}:6143/tasks/{Task ID}

# Fetch the result of a Stopped task
curl -i {Node IP or hostname}:6143/tasks/{Task ID}/result

# Delete the task from the node's history
curl -iX {Node IP or hostname}:6143/tasks/{Task ID}
```

