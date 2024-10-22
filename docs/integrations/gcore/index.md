---
title: 'GCore'
date: 2024-10-22T10:38:15-04:00
draft: false
weight: 1
---

GCore provided supports their [managed DNS](https://gcore.com/dns) offer. First, you'll need to [create an API token](https://portal.gcore.com/accounts/profile/api-tokens) and then store the value of the generated token into a secret.

```sh
API_TOKEN="GCORE-generated-secret"
kubectl create secrets generic gcore-secrets \
  --namespace phonebook-system \
  --from-literal=apiToken=${API_TOKEN}
```

```yaml
apiVersion: se.quencer.io/v1alpha1
kind: DNSIntegration
metadata:
  name: gcore
spec:
  provider:
    name: gcore
  zones:
    - mydomain.com
  secretRef:
    name: gcore-secrets
    keys:
      - key: "apiToken"
        name: "GCORE_API_TOKEN"
```

## Deploying

Now you can deploy with the normal command:
```
helm upgrade --install phonebook phonebook/phonebook \
  --namespace phonebook-system \
  --create-namespace \
```
