# configmap-attacher

## Overview
This container was built to provide the ability to work with revisioned configmaps while performing Canary deployment with Argo Rollouts on Kubernetes.
**It supports only Rollout objects and not deployment objects with workloadRef**

[This Job is part of the following Helm chart](https://github.com/liorfranko/base-app)
### Configmaps and Canary Deployment:
When performing a Canary deployment, there will be a new ReplicaSet and an old ReplicaSet.
Based on the Canary strategy, pods will be migrated from the old ReplicaSet to the new ReplicaSet.
By default, when a Canary deployment is triggered due to a change in a configmap [link](https://helm.sh/docs/howto/charts_tips_and_tricks/#automatically-roll-deployments), the old configmap will be replaced with the new one, and Pods that will be migrated to the new ReplicaSet will load the new configmap.
### The problem:
During the Canary rollout, one of the pods from the old ReplicaSet is recreated or restarted due to infra-related causes like Spot replacement; that pod will load the new configmap although it's located on the old ReplicaSet.
### The solution
On every deploy, attach the revisioned configmaps to the new ReplicaSets, and utilize the K8S garbage collector to delete old configmaps.

## CLI
* `-rollout` - Rollout that will be the ownerReference
* `-namespace` - The namespace of the rollout and configmaps
* `-configmaps` - Configmaps to add the ownerReference, for multiple configmaps use ',' as a separator

## Environment Variables
| Variable name | Description | Default | Required |
| --- | --- | --- | --- |
| IS_IN_CLUSTER | Whether to use in cluster communication or to look for a kubeconfig in home directory | true | N/A |
| LOG_LEVEL | Logger's log granularity (debug, info, warn, error, fatal, panic) | info |N/A |
| VERSION | For logging audit, please add the version of the configmap-attacher | None | true |
## Permissions
To make the configmap-attacher work on any namespace, it's better to deploy it in kube-system with ClusterRole permissions; you can use the following:

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
## Job example:
```
apiVersion: batch/v1
kind: Job
metadata:
  namespace: kube-system
  annotations:
    argocd.argoproj.io/hook: Sync
  name: configmap-attacher-job
spec:
  backoffLimit: 3
  template:
    metadata:
      name: configmap-attacher-job
      annotations:
    spec:
      containers:
      - name: configmap-attacher-job
        args: ["-rollout", "<Rollout Name>", "-namespace", "<Namespace>", "-configmaps", <Configmap name>]
        image: quay.io/liorfranko/configmap-attacher:1.0.1
        resources:
          limits:
            cpu: 0.1
            memory: 100Mi
          requests:
            cpu: 0.1
            memory: 100Mi
        env:
          - name: VERSION
            value: "1.0.1"
      restartPolicy: OnFailure
      serviceAccountName: configmap-attacher-job
```
