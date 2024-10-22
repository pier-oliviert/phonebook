---
title: 'deSEC'
date: 2024-09-27T10:38:15-04:00
draft: false
weight: 1
---

## Example DNSIntegration records

Create a DNSIntegration to start using your deSEC zone with Phonebook

You will need to create a deSEC API token with relevant permissions.

```yaml
apiVersion: se.quencer.io/v1alpha1
kind: DNSIntegration
metadata:
  name: desec
spec:
  provider:
    name: desec
  zones:
    - mydomain.com
  secretRef:
    name: desec-secrets
    keys:
      - name: "DESEC_TOKEN"
        key: 'sometokenfromdesechere'
```

If you wish to use environment variables over secrets:
```yaml
apiVersion: se.quencer.io/v1alpha1
kind: DNSIntegration
metadata:
  name: desec
spec:
  provider:
    name: desec
  zones:
    - mydomain.com
  env:
    - name: DESEC_TOKEN
      value: 'sometokenfromdesechere'
```

## Deploying

Now you can deploy with the normal command:
```
helm upgrade --install phonebook phonebook/phonebook \
  --namespace phonebook-system \
  --create-namespace \
  --values values.yaml
```
