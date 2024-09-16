# Phonebook: Manage DNS Record in Kubernetes

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
  Name: mysubdomain
  Targets:
    - 127.0.0.1
```

### Features

- Full DNS Record lifecycle
- Each DNS Record owns a status that reflects the current state of the DNS Record as well errors, if any.
- Support all DNS Record Types (A, AAAA, TXT, CNAME, etc.)
- Support cloud provider specific properties 
- Can be managed like any other Kubernetes resources, (k9s, Argo, Github Actions, etc.)
- Great visibility for each DNS Record

### Supported providers

Here's a list of all supported providers. If you need a provider that isn't yet supported, create a new [issue](https://github.com/pier-oliviert/phonebook/issues/new).

|Provider|
|--|
|[AWS](./docs/providers/aws.md)|
|[Cloudflare](./docs/providers/cloudflare.md)|

### Get Started

Check the [Get Started](./GET_STARTED.md) page to learn how to install Phonebook in your Kubernetes cluster.

### Special thanks

This project was built out of need, but I also want to give a special thanks to [external-dns](https://github.com/kubernetes-sigs/external-dns) as that project was a huge inspiration for Phonebook. A lot of the ideas here stem from my usage of external-dns over the years. I have nothing but respect for that project.
