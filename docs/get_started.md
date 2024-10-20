---
title: 'Get Started'
date: 2024-09-20T10:38:15-04:00
draft: false
weight: 1
---

The helm chart is the official way to install Phonebook in your cluster.

1. **Add helm repo**

```sh
helm repo add phonebook https://pier-oliviert.github.io/phonebook/ --force-update
```

2. **Install Phonebook**
```sh
helm upgrade --install phonebook phonebook/phonebook \
  --namespace phonebook-system \
  --create-namespace \
  --values values.yaml
```

3. **Create a DNSIntegration**

Phonebook requires some user values about the DNS provider you want to use to successfully run. Refer to the [integrations]({{< ref "/integrations" >}}) section to learn how to configure your `values.yaml` file.

The DNSIntegration is what configure one of the supported providers
