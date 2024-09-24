---
title: 'Phonebook'
date: 2024-09-20T10:38:15-04:00
draft: false
cascade:
  type: docs
---

Phonebook is a Kubernetes Operator that lets you manage DNS Record like any other resource in Kubernetes -- Deployments, Services, etc. You can safely create and delete DNS Record from `kubectl` and it will do the right thing.

```yaml
# This will create a new `A` record `helloworld.gotta-be-kidding.com` pointing
# at `127.0.0.1``
apiVersion: se.quencer.io/v1alpha1
kind: DNSRecord
metadata:
  name: dnsrecord-sample
  namespace: phonebook-system
spec:
  zone: gotta-be-kidding.com
  recordType: A
  name: helloworld
  targets:
    - 127.0.0.1
```

![A DNS Record](status.png)

## Features

- Only manage DNS Record that are presents as DNSRecord in the cluster
- Manage DNS Record like any other resources (Create/Delete)
- Support all DNS Record Types (A, AAAA, TXT, CNAME, etc.)
- Support cloud provider specific properties 
- Proper error handling per DNS Record

