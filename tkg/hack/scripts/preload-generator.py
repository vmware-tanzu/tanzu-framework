#!/usr/bin/env python

import os

from kubernetes import client, config

config.load_kube_config()

v1 = client.CoreV1Api()

loaded_images = []

ret = v1.list_pod_for_all_namespaces(watch=False)
for i in ret.items:
    for images in i.spec.containers:
        if i.metadata.namespace in ['kube-system', "kube-public", "kube-node-lease", "local-path-storage"]:
            next
        else:
            loaded_images.append(images.image)

loaded_images = list(set(loaded_images)))

docker_pull_images=map(lambda image: "docker pull " + image, loaded_images)
kind_load_images = map(lambda image: "kind load docker-image " + image, loaded_images)

print("\n".join(docker_pull_images))
print("\n".join(kind_load_images))
