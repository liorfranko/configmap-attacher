# configmap-attacher

## Overview
This container was built to solve the problem of revisioned configmaps on Kubernetes for Canary deployments, it discover and attaches configmaps to new replicasets created by Argo Rollout.

**It supports only Rollout objects and not deployment objects with workloadRef**

## CLI
* `-rollout` - The rollout name
* `-namespace` - The namespace of the rollout and configmaps
* `-configmaps` - One or more configmaps, for multiple configmaps use "," as a seperator

## Environmet Variables
| Variable name | Description | Default | Required |
| --- | --- | --- | --- |
| IS_IN_CLUSTER | Whether to use in cluster communication or to look for a kubeconfig in home directory | true | N/A |
| LOG_LEVEL | Logger's log granularity (debug, info, warn, error, fatal, panic) | info |N/A |
| VERSION | For logging audit please add the version of the configmap-attacher | None | true |
## Permissions
To make the configmap-attacher work on any namespace, it's better to deploy it in kube-system with ClusterRole permissions, you can use the following:
```
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: configmap-attacher-job
rules:
- apiGroups:
  - argoproj.io
  resources:
  - rollouts
  - rollouts/status
  - rollouts/finalizers
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - apps
  resources:
  - replicasets
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
  - patch
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: configmap-attacher-job
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: configmap-attacher-job
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: configmap-attacher-job
subjects:
- kind: ServiceAccount
  name: configmap-attacher-job
  namespace: kube-system

```
