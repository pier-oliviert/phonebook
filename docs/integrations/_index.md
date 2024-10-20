---
title: 'Integrations'
date: 2024-09-20T10:38:15-04:00
draft: false
cascade:
  type: docs
weight: 3
---

Phonebook's integration with providers exists through the `DNSIntegration` cluster-scope CRD. Each integration will run its own deployment that manages `DNSRecord` under it's zone authority. To give you a better idea of how this work, imagine a cluster where the 2 following integrations are created.

```yaml
apiVersion: se.quencer.io/v1alpha1
kind: DNSIntegration
metadata:
  name: cloudflare-demo
spec:
  provider:
    name: cloudflare
  zones:
    - mydomain.com
  secretRef:
    name: cloudflare-secrets
    keys:
      - key: CF_API_TOKEN
        name: CF_API_TOKEN
      - key: CF_ZONE_ID
        name: CF_ZONE_ID
```

```yaml
apiVersion: se.quencer.io/v1alpha1
kind: DNSIntegration
metadata:
  name: azure-demo
spec:
  provider:
    name: azure
  zones:
    - myotherdomain.com
  secretRef:
    name: azure-secrets
    keys:
      - name: "AZURE_ZONE_NAME"
        key: "zoneName"
      - name: "AZURE_RESOURCE_GROUP"
        key: "rgName"
      - name: "AZURE_SUBSCRIPTION_ID"
        key: "subId"
      - name: "AZURE_TENANT_ID"
        key: "tenantId"
      - name: "AZURE_CLIENT_ID"
        key: "clientId"
      - name: "AZURE_CLIENT_SECRET"
        key: "clientSecret"
```

Any DNSRecord created with the zone `mydomain.com` would be handled by the `cloudflare-demo` integration. If you'd create a DNSRecord with `myotherdomain.com` as the zone, Azure will be used. 

```yaml
# This will create a new `A` record `helloworld.mydomain.com` pointing
# at `127.0.0.1` using `cloudflare-demo` as the integration
apiVersion: se.quencer.io/v1alpha1
kind: DNSRecord
metadata:
  name: dnsrecord-sample
  namespace: phonebook-system
spec:
  zone: mydomain.com
  recordType: A
  name: helloworld
  targets:
    - 127.0.0.1
    - 127.0.0.2 # If provider supports multi-target    
```

## Split-Horizon DNS

Alternatively, if you want to do [split-horizon DNS](https://en.wikipedia.org/wiki/Split-horizon_DNS), both integrations would share the same zone. Let's use the same `mydomain.com` and configure both cloudflare and azure to use it.

```yaml
apiVersion: se.quencer.io/v1alpha1
kind: DNSIntegration
metadata:
  name: cloudflare-demo
spec:
  provider:
    name: cloudflare
  zones:
    - mydomain.com
  secretRef:
    name: cloudflare-secrets
    keys:
      - key: CF_API_TOKEN
        name: CF_API_TOKEN
      - key: CF_ZONE_ID
        name: CF_ZONE_ID
```

```yaml
apiVersion: se.quencer.io/v1alpha1
kind: DNSIntegration
metadata:
  name: azure-demo
spec:
  provider:
    name: azure
  zones:
    - mydomain.com # Same as cloudflare-demo
  secretRef:
    name: azure-secrets
    keys:
      - name: "AZURE_ZONE_NAME"
        key: "zoneName"
      - name: "AZURE_RESOURCE_GROUP"
        key: "rgName"
      - name: "AZURE_SUBSCRIPTION_ID"
        key: "subId"
      - name: "AZURE_TENANT_ID"
        key: "tenantId"
      - name: "AZURE_CLIENT_ID"
        key: "clientId"
      - name: "AZURE_CLIENT_SECRET"
        key: "clientSecret"
```

Now, you'll want to have different values for the same DNS Record. You can do this by using the optional `integration` field in the `DNSRecord`.

```yaml
# Use azure for this record
apiVersion: se.quencer.io/v1alpha1
kind: DNSRecord
metadata:
  name: hello-azure
  namespace: phonebook-system
spec:
  zone: mydomain.com
  recordType: A
  name: helloworld
  targets:
    - 127.0.0.1
  integration: azure-demo
```

```yaml
# Use cloudflare for this record
apiVersion: se.quencer.io/v1alpha1
kind: DNSRecord
metadata:
  name: hello-cloudflare 
  namespace: phonebook-system
spec:
  zone: mydomain.com
  recordType: A
  name: helloworld
  targets:
    - 127.0.0.5
  integration: cloudflare-demo
```

