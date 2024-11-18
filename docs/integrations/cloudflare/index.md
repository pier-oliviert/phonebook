---
title: 'Cloudflare'
date: 2024-09-20T10:38:15-04:00
draft: false
weight: 1
---

```yaml
apiVersion: se.quencer.io/v1alpha1
kind: DNSIntegration
metadata:
  name: cloudflare
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

To use Cloudflare as a provider, you'll need to create an API token on their site and create a secret in your Kubernetes cluster. Phonebook expects the secret to live in the **same namespace as the one running Phonebook's controller**.

```sh
kubectl create secrets generic cloudflare-secrets \
  --namespace phonebook-system \
  --from-literal=apiToken=${API_TOKEN} \
  --from-literal=zoneId=${ZONE_ID} \
```

&nbsp;

### API Token

The API Token can be created by going to your [Cloudflare's profile page](https://dash.cloudflare.com/profile/api-tokens). Create a new token that will include the two permissions:

1. `Zone.DNS` for `All Zones`
2. `Account.Cloudflare Tunnel` for `All Account`

> ![Cloudflare's token page](./token-page.png)

It's possible to narrow down the zones and accounts to the specific one you want to use, but this is an exercise to the user. Once the API Token is created, you'll need to create a secrets, like the one at the top of this page, that includes your API token as well as the zone id.

&nbsp;

### Zone ID

![Domain's page with Zone and Account IDs](profile-page.png)

### Tags and comments

Cloudflare has support for comment and tags for each record. If you want to create a DNS Record that includes both, you can use the `Properties` field of the DNSRecord:

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
    - 127.0.0.2 # If provider supports multi-target    
  properties:
    comment: "A Comment for this record"
    tags: "phonebook;development;user-1"
```

Tags are delimited by the semi-colon character(`;`). You can specify as many as you'd like.
