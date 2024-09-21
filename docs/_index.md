---
title: 'Phonebook'
date: 2024-09-20T10:38:15-04:00
draft: false
cascade:
  types: doc
---

Manage your DNS Record in Kubernetes like you manage everything else.

```yaml
# This will create a new `A` record `mysubdomain.mytestdomain.com` pointing
# at `127.0.0.1``
apiVersion: se.quencer.io/v1alpha1
kind: DNSRecord
metadata:
  name: dnsrecord-sample
  namespace: phonebook-system
spec:
  zone: mytestdomain.com
  recordType: A
  name: mysubdomain
  targets:
    - 127.0.0.1
```
