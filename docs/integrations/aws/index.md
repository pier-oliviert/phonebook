---
title: 'AWS'
date: 2024-09-20T10:38:15-04:00
draft: false
weight: 1
---

## Zone ID

The AWS Zone ID needs to be specified in your `values.yaml`. The Zone ID needs to point to the domain you want to manage. Currently, only 1 domain can be managed by Phonebook.

## Authentication
You have two options when you configure your AWS provider. Depending on your setup, one might be more suited to your need than the other.

- IAM Role bound to a service account
- User supplied credentials

### IAM Role bound to a service account

This option is the recommended one if your Kubernetes cluster supports it. Most of the configuration happens on the AWS control panel and the changes to your helm chart is minimal. Moreover, you won't have to store any credentials at all on K8s as it will all be taken care of by EKS.

{{< callout type="warning" >}}
You should already have an EKS cluster running in your account, with a running node group. The EKS Cluster should also be configured to use **EKS API** as an authentication mode.
{{< /callout >}}

First, you'll need to add an annotation to the serviceAccount that Phonebook uses to run the Provider's deployment.

```yaml
serviceAccount:
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::1111111111:role/Phonebook-ServiceAccount
```

Then, you can create a DNSIntegration that is configured with the `AWS_ZONE_ID`.
```yaml
apiVersion: se.quencer.io/v1alpha1
kind: DNSIntegration
metadata:
  name: aws
spec:
  provider:
    name: aws
  zones:
    - mydomain.com
  env:
    - name: AWS_ZONE_ID
      value: Z1111111111111
```


#### Configure OIDC (OpenID Connect)

This section will guide you through configuring your [EKS](https://aws.amazon.com/eks/) cluster to run with Phonebook. It will focus on configuring Phonebook's [`serviceAccount`](https://kubernetes.io/docs/concepts/security/service-accounts/) to allow its controller to make changes to [Route53](https://aws.amazon.com/route53/) on your behalf.

Once the ServiceAccount is fully configure, Phonebook should be able to make changes to your DNS records by automatically authenticating to AWS using the right permissions as set here; No access token/secret token will be required to be set by you.


EKS comes with a few defaults, but OIDC is not fully configured out of the box. Documentation is available on [AWS](https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html) that shows how to connect your EKS cluster to an Identity Provider in your IAM console.

Once your Identity Provider is configured, you'll be ready to create a role to use it. Keep your OIDC Provider URL close by, you'll need it in the following section

> ![EKS cluster detail page](./cluster-page.png)

#### Create a role for your Service Account

OIDC is the bridge that can connect IAM to your EKS cluster. To make it possible for Phonebook to make changes to Route53, you'll need to create a role that you'll use as annotations with your Service Account. 

{{< callout type="info" >}}
This section will use a fake OIDC Provider URL based on the screenshot above:

- OIDC Provider URL: `https://oidc.eks.us-east-2.amazonaws.com/id/F1A5247B9AAAAAAAAAAAAA06364EC072201D`
- OIDC ID: `F1A5247B9AAAAAAAAAAAAA06364EC072201D`

You'll need to replace those values with the ones you have.
{{< /callout >}}

First, create a new Role. For Trusted Entity, select **Custom trust policy**. In the textbox that should have appeared, replace the content with this template:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Federated": "arn:aws:iam::${YOUR_ACCOUNT_ID}:oidc-provider/oidc.eks.${YOUR_REGION}.amazonaws.com/id/${OIDC_ID}"
            },
            "Action": "sts:AssumeRoleWithWebIdentity",
            "Condition": {
                "StringEquals": {
                    "oidc.eks.us-east-2.amazonaws.com/id/${OIDC_ID}:aud": "sts.amazonaws.com",
                    "oidc.eks.us-east-2.amazonaws.com/id/${OIDC_ID}:sub": "system:serviceaccount:${PHONEBOOK_NAMESPACE}:phonebook-providers"
                }
            }
        }
    ]
}
```

|${Variable}|Description|
|--|--|
|YOUR_ACCOUNT_ID|Your AWS Account ID. This is usually a number with 12 digits.|
|YOUR_REGION|The AWS Region for your EKS Cluster. It is part of the OIDC Provider URL. In the example above, the URL is `https://oidc.eks.us-east-2.amazonaws.com/...` which means the region, in this example, is `us-east-2`|
|OIDC_ID|In the example above, the OIDC_ID is `F1A5247B9AAAAAAAAAAAAA06364EC072201D`|
|PHONEBOOK_NAMESPACE|The namespace, in Kubernetes, that phonebook runs in. By default, this value is `phonebook-system`. If you changed it, you'll need to use the same value here too.|


#### Define a policy for the new role

A policy, in AWS, represents what service does the role has access to. Phonebook only needs access to a few of Route53's action. Most likely, you'll need to create a new custom policy for your Role as part of the Role wizard. You can name it whatever you want, it'll only be used by the Role here and won't need to be referenced anywhere.

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "route53:ChangeResourceRecordSets"
            ],
            "Resource": [
                "arn:aws:route53:::hostedzone/*"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "route53:ListHostedZones",
                "route53:ListResourceRecordSets",
                "route53:ListTagsForResource"
            ],
            "Resource": [
                "*"
            ]
        }
    ]
}
```

#### Add the Role as annotations to Phonebook's service account

The last piece of the puzzle is to add an annotation to Phonebook's Service account so EKS can elevate this account. In the Role's detail page, you should have the **ARN** for that role. It should look something like this: `arn:aws:iam::1111111111:role/Phonebook-ServiceAccount`

Modify your `values.yaml` to include the Role ARN as an annotations.

```yaml
serviceAccount:
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::1111111111:role/Phonebook-ServiceAccount
```

### User supplied credentials

Although simpler at face value, supplied credentials requires you to manage rotation and make sure that secrets are present in the cluster before using them. Since Phonebook uses the official Go SDK for AWS, you can refer to AWS's [official documentation](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/) if you want to know more.

```sh
kubectl create secrets generic aws-secret \
  --namespace phonebook-system \
  --from-literal=accessKeyID=${ACCESS_KEY} \
  --from-literal=secretAccessKey=${SECRET_KEY} \
  --from-literal=sessionToken=${SESSION_TOKEN} \
```

Once created, you can create the DNSIntegration that will configure a provider with the secrets you generated.

```yaml
apiVersion: se.quencer.io/v1alpha1
kind: DNSIntegration
metadata:
  name: aws
spec:
  provider:
    name: aws
  zones:
    - mydomain.com
  secretRef:
    name: aws-secret
    keys:
      - name: "AWS_ACCESS_KEY_ID"
        key: "accessKeyID"
      - name: "AWS_SECRET_ACCESS_KEY"
        key: "secretAccessKey"
      - name: "AWS_SESSION_TOKEN"
        key: "sessionToken"
```
