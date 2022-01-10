#!/usr/bin/env python3
import os
import optparse
import sys
import json
import pprint
usage = "Usage: %prog [options]"
parser = optparse.OptionParser(usage)
parser.add_option("--configmap", action="store", dest="configmap", help="Configmap to add the ownerReference - mandatory")
parser.add_option("--rollout", action="store", dest="rollout", help="Rollout that will be the ownerReference - mandatory")
parser.add_option("--namespace", action="store", dest="namespace", help="The namespace of the rollout and configmap - mandatory")


options, _ = parser.parse_args()

if options.configmap is None or options.rollout is None or options.namespace is None:
    parser.print_help()
    sys.exit()

stream = os.popen("kubectl -n {} get rollout {} -o json".format(options.namespace, options.rollout))
output = stream.read()
print (output)
y = json.loads(output)
# print (y['spec'])
print ('Old replicaset is: {}'.format(y['status']['stableRS']))
print ('New replicaset is: {}'.format(y['status']['currentPodHash']))
newRS = y['status']['currentPodHash']
stream = os.popen("kubectl -n {} get replicasets.apps {}-{} -o json".format(options.namespace, options.rollout, y['status']['currentPodHash']))
output = stream.read()
y = json.loads(output)
print ('Replicaset UID is: {}'.format(y['metadata']['uid']))
uid = y['metadata']['uid']
stream = os.popen("kubectl -n {} get configmap {} -o json".format(options.namespace, options.configmap))
output = stream.read()
y = json.loads(output)
print ('Configmap is: {}'.format(y))
command = "kubectl -n {} patch configmap {} --type merge -p '{{\"metadata\":{{\"ownerReferences\":[{{\"apiVersion\":\"apps/v1\",\"blockOwnerDeletion\":true,\"controller\":true,\"kind\":\"ReplicaSet\",\"name\":\"{}-{}\",\"uid\":\"{}\"}}]}}}}'".format(options.namespace, options.configmap, options.rollout,newRS, uid)
print (command)
stream = os.popen(command)
output = stream.read()
print (output)
exit()
