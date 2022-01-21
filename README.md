# configmap-attacher

## Overview
This container was built to solve the problem of revisioned configmaps on Kubernetes for Canary deployments, it discover and attaches configmaps to new replicasets created by Argo Rollout.

**It supports only Rollout objects and not deployment objects with workloadRef**

## CLI
* `-rollout` - The rollout name
* `-namespace` - The namespace of the rollout and configmaps
* `-configmaps` - One or more configmaps, for multiple configmaps use "," as a seperator

## Permissions
The container needs the following permissions to work:
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
# replicaset access needed for managing ReplicaSets
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
