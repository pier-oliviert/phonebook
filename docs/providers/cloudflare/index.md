---
title: 'Cloudflare'
date: 2024-09-20T10:38:15-04:00
draft: false
---

```yaml
# values.yaml
controller:
  env:
    PHONEBOOK_PROVIDER: cloudflare
  providerSecrets: cloudflare-secrets
```

```sh
kubectl create secrets generic cloudflare-secrets \
  --namespace phonebook-system \
  --from-literal=CF_API_TOKEN=${API_TOKEN} \
  --from-literal=CF_ZONE_ID=${ZONE_ID} \
```

To use Cloudflare as a provider, you'll need to create an API token on their site and create a secret in your Kubernetes cluster. Phonebook expects the secret to live in the **same namespace as the one running Phonebook's controller**.

&nbsp;

### API Token

The API Token can be created by going to your [Cloudflare's profile page](https://dash.cloudflare.com/profile/api-tokens). Create a new token that will include the two permissions:

1. `Zone.DNS` for `All Zones`
2. `Account.Cloudflare Tunnel` for `All Account`

> ![Cloudflare's token page](./token-page.png)

It's possible to narrow down the zones and accounts to the specific one you want to use, but this is an exercise to the user. Once the API Token is created, you'll need to add it to the cluster, using the secret's name `cloudflare-api-token` as defined in the example above.

```sh
kubectl create secret generic cloudflare-api-token \
  --from-literal=CF_API_TOKEN=$(MY_CLOUDFLARE_API_TOKEN) \
  --from-literal=CF_ZONE_ID=$(MY_ZONE_ID) \
  --namespace phonebook-system
```

&nbsp;

### Account IDs

![Domain's page with Zone and Account IDs](profile-page.png)


