# Autoscaler support using the TKG CLI

**Goal:** Users should be able to deploy Autoscaler automatically for a workload cluster if they want their worker nodes to be scaled up/down automatically.

Autoscaler is deactivated by default when creating workload clusters. Users can enable autoscaler using the `ENABLE_AUTOSCALER: true` config variable.

There are multiple configuration options for users to be able to configure the autoscaler deployment. And they can be configured using the TKG config file or via env variables.

| Config Variable | Default | Description |
| --- | --- | --- |
| AUTOSCALER_MAX_NODES_TOTAL | 0 | Maximum number of nodes in all node groups or machine deployments. Cluster autoscaler will not grow the cluster beyond this number.  If it is defaulted to 0, this configuration is not respected anymore and autoscaler will then only respect min/max values for each Machine deployment |
| AUTOSCALER_SCALE_DOWN_DELAY_AFTER_ADD | 10m | How long after scale up that scale down evaluation resumes |
| AUTOSCALER_SCALE_DOWN_DELAY_AFTER_DELETE | 10s | How long after node deletion that scale down evaluation resumes |
| AUTOSCALER_SCALE_DOWN_DELAY_AFTER_FAILURE | 3m | How long after scale down failure that scale down evaluation resumes |
| AUTOSCALER_SCALE_DOWN_UNNEEDED_TIME | 10m | How long a node should be unneeded before it is eligible for scale down |
| AUTOSCALER_MAX_NODE_PROVISION_TIME | 15m | Maximum time autoscaler waits for node to be provisioned |
| AUTOSCALER_MIN_SIZE_0 | Defaults to the value of WORKER_MACHINE_COUNT_0 if not set | Minimum number of nodes the autoscaler will scale down workers in the first AZ to |
| AUTOSCALER_MAX_SIZE_0 | Defaults to the value of WORKER_MACHINE_COUNT_0 if not set | Maximum number of nodes the autoscaler will scale up workers in the first AZ to |
| AUTOSCALER_MIN_SIZE_1 | Defaults to the value of WORKER_MACHINE_COUNT_1 if not set | Minimum number of nodes the autoscaler will scale down workers in the second AZ to (incase of multiple machine deployments) |
| AUTOSCALER_MAX_SIZE_1 | Defaults to the value of WORKER_MACHINE_COUNT_1 if not set | Maximum number of nodes the autoscaler will scale up workers in the second AZ to (incase of multiple machine deployments) |
| AUTOSCALER_MIN_SIZE_2 | Defaults to the value of WORKER_MACHINE_COUNT_2 if not set | Minimum number of nodes the autoscaler will scale down workers in the third AZ to (incase of multiple machine deployments) |
| AUTOSCALER_MAX_SIZE_2 | Defaults to the value of WORKER_MACHINE_COUNT_2 if not set | Maximum number of nodes the autoscaler will scale up workers in the third AZ to (incase of multiple machine deployments) |


More details about autoscaler and defaults here - https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/FAQ.md