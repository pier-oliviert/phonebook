# Phonebook: Manage DNS Record in Kubernetes
[![Tests](https://github.com/pier-oliviert/phonebook/actions/workflows/test.yaml/badge.svg)](https://github.com/pier-oliviert/phonebook/actions/workflows/test.yaml)

Phonebook is an [operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) that helps you manage DNS Record for your cloud provider from within Kubernetes. Using custom resource definitions (CRDs), you can build DNS records in a same manner you would create other resources with Kubernetes.

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
  ttl: 60
  targets:
    - 127.0.0.1
    - 127.0.0.2 # If provider supports multi-target
```

### Features

- Only manage DNS Record that are presents as DNSRecord in the cluster
- Manage DNS Record like any other resources (Create/Delete)
- Support all DNS Record Types (A, AAAA, TXT, CNAME, etc.)
- Support cloud provider specific properties 
- Proper error handling per DNS Record
- Allows specifying TTL
- Allows multiple targets on providers with multi support (Azure, AWS)
### Supported providers

Here's a list of all supported providers. If you need a provider that isn't yet supported, create a new [issue](https://github.com/pier-oliviert/phonebook/issues/new).

|||||
|--|--|--|--|
|[AWS](https://pier-oliviert.github.io/phonebook/providers/aws/)|[Cloudflare](https://pier-oliviert.github.io/phonebook/providers/cloudflare/)|[Azure](https://pier-oliviert.github.io/phonebook/providers/azure/)|[deSEC](https://pier-oliviert.github.io/phonebook/providers/desec/)

### Get Started

The [documentation](https://pier-oliviert.github.io/phonebook/) has all the information for you to get started with Phonebook.

### Special thanks

This project was built out of need, but I also want to give a special thanks to [external-dns](https://github.com/kubernetes-sigs/external-dns) as that project was a huge inspiration for Phonebook. A lot of the ideas here stem from my usage of external-dns over the years. I have nothing but respect for that project.
