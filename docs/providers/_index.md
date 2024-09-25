---
title: 'Providers'
date: 2024-09-20T10:38:15-04:00
draft: false
cascade:
  type: docs
weight: 3
---

This is the complete list of all DNS providers supported by Phonebook. Each provider have different requirements, so please read the section below that is associated with the provider you want to use.


## Environment Variables

Phonebook come with an wrapper for values you need to provide to your provider. For values that aren't sensitive, you can set them up like any other values. For secrets and other sensitive information, you can store those in a secret and let Phonebook use this secret.

The only requirement for secrets is that names for each secret need to use the same format as environment variables. For instance, if you want to use Cloudflare's Provider, you'll needto set a value for `CF_API_TOKEN`.

The non-secret way would be to store the value in the environment variable directly:

```yaml
controller:
  env:
    - name: PHONEBOOK_PROVIDER
      value: cloudflare
      name: CF_API_TOKEN
      value: MySecretToken
```

This is the simplest, but also leave your token in plain text within your cluster. Anyone can see the value just by looking at the deployment.

If you want to avoid showing the content of the API Token in plain text, you'll want to use a Kubernetes secret:
```yaml
controller:
  env:
    - name: PHONEBOOK_PROVIDER
      value: cloudflare
  providerSecrets: cloudflare-secrets
```

The `providerSecrets` key tells Phonebook to mount the content of the secret as read-only files *inside* the pod. Each key/pair in the secret will be mounted as files. Given the example above, you'd need to create a secrets with the same name.

```sh
kubectl create secrets generic cloudflare-secrets \
  --namespace phonebook-system \
  --from-literal=CF_API_TOKEN=MySecretToken \
  --from-literal=CF_ZONE_ID=${ZONE_ID} \
```

Notice that the key `CF_API_TOKEN` is the same as the environmnent variable. This is the only requirement for using secrets, the keys needs to be all capitalized, like environment variables. Phonebook [mounts the secret through a volumne](https://kubernetes.io/docs/tasks/inject-data-application/distribute-credentials-secure/#create-a-pod-that-has-access-to-the-secret-data-through-a-volume):

|Name|Value|Path|
|--|--|--|
|CF_API_TOKEN|MySecretToken|/var/run/configs/provider/CF_API_TOKEN|
|CF_ZONE_ID|${ZONE_ID}|/var/run/configs/provider/CF_ZONE_ID|
