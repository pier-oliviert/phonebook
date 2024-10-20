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
  --create-namespace
```

3. **Create a DNSIntegration**

Phonebook requires at least one DNSIntegration to work. These integrations can be seen as the glue between a DNS Provider (aws, cloudflare, azure, etc.) and a DNS Record created with Phonebook. Since each of the integration requires different settings and values, please refer to the [integrations]({{< ref "/integrations" >}}) section to learn how to create a DNSIntegration based off the provider you want to use.
