# Overview

## POC that enables mirroring using remote worker nodes (bare-metal or VM's)

**NB** This is still a WIP the cli and suggested implementation is continuously under change

## Usage

Before mirroring we need to build the binary

```
make build
```
Once this is done update the inventory list (ansible/workers)

Change the variables accordingly (ansible/roles/manage/vars/main.yaml)

Change the SERVER_PORT and CALLBACK_URL in run.sh 

We can now deploy the relevant artifacts

```


```

For the mirror to use this feature use the following cli (in v2 of oc-mirror)

``` bash

build/mirror --config isc.yaml file://test --loglevel trace --distributed-workers inventory

# imagesetconfig used

---
apiVersion: mirror.openshift.io/v1alpha2
kind: ImageSetConfiguration
mirror:
  operators:
    - catalog: registry.redhat.io/redhat/redhat-operator-index:v4.13
      packages:
      - name: node-observability-operator
      - name: aws-load-balancer-operator
      - name: 3scale-operator
```
