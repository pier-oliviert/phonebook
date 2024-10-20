---
title: "SSL Cert with Cert-Manager"
date: 2024-09-30T20:42:38.603Z
draft: false
cascade:
  type: docs
---

One of the benefit of using Phonebook is that it comes with full support for [Let's Encrypt](https://letsencrypt.org/) DNS-01 Challenge with Cert-Manager. What does that mean for you?

It means you can create SSL Certificate for any domain you own, **including wildcards Certificates**. Those certificates can also be dynamically [using](https://cert-manager.io/docs/usage/certificate/) [cert-manager's](https://cert-manager.io/docs/usage/ingress/) [annotations](https://cert-manager.io/docs/usage/gateway/).

## Configure Cert-Manager

You'll obviously need to have cert-manager running in your cluster. If you need help to install it, their documentation is pretty thorough: [https://cert-manager.io/docs/installation/](https://cert-manager.io/docs/installation/). Once that's done, you'll need to configure a new Issuer:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: phonebook-acme-issuer
spec:
  acme:
    email: "youremail@exmaple.com"
    server: "https://acme-v02.api.letsencrypt.org/directory"
    privateKeySecretRef:
      name: acme-issuer
    solvers:
      - dns01:
          webhook:
            groupName: phonebook.se.quencer.io
            solverName: solver

```

The `email` field needs to be set to an email adddress **you own**. Once it is set, save your yaml file (ie. `issuer.yaml`) and create the issuer:

```bash
kubectl create -f issuer.yaml
```

## Enable DNS-01 Solver on Phonebook
While the Issuer is fully configured at this point, Phonebook, by default, doesn't have the DNS-01 Solver running. To enable it, you can update your Helm installation:

```bash
helm upgrade --install phonebook phonebook/phonebook \
  --namespace phonebook-system \
  --create-namespace \
  --set solver.enabled=true
```

Once this call returns, Phonebook's controller should restart and if you inspect your deployment, you should see that the controller now runs with an extra argument (`--solver`). You should now be ready to create SSL certificate using cert-manager with Let's Encrypt.

## Examples
These examples are copies of examples you can find in Cert-Manager's docuemntation pages. The Issuer was changed to the one created above to give you an idea of how you can make it work for you.

### Ingress Annotations

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    cert-manager.io/cluster-issuer: phonebook-acme-issuer
  name: myIngress
  namespace: myIngress
spec:
  rules:
  - host: example.com
    http:
      paths:
      - pathType: Prefix
        path: /
        backend:
          service:
            name: myservice
            port:
              number: 80
  tls: # < placing a host in the TLS config will determine what ends up in the cert's subjectAltNames
  - hosts:
    - example.com
    secretName: myingress-cert # < cert-manager will store the created certificate in this secret.
```

### Certificate
```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: example-com
  namespace: phonebook-system
spec:
  secretName: example-com-tls

  privateKey:
    algorithm: RSA
    encoding: PKCS1
    size: 2048

  duration: 2160h # 90d
  renewBefore: 360h # 15d

  isCA: false
  usages:
    - server auth
    - client auth

  subject:
    organizations:
      - cert-manager

  commonName: mydomain.com
  dnsNames:
    - "mydomain.com
    - "*.mydomain.com"

  # Issuer references are always required.
  issuerRef:
    name: phonebook-acme-issuer
    kind: ClusterIssuer
    group: cert-manager.io
```
