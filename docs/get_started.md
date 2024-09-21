---
title: 'Get Started'
date: 2024-09-20T10:38:15-04:00
draft: false
weight: 1
---

The helm chart is the official way to install Phonebook in your cluster.

```sh
helm upgrade --install phonebook $TODO_URL \
  --namespace phonebook-system \
  --create-namespace \
  --values values.yaml
```

The `values.yaml` is your own file that you need to configure to use the provider you want to use.
