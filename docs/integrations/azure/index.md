---
title: 'Azure'
date: 2024-09-27T10:38:15-04:00
draft: false
weight: 1
---

## Obtaining Access to the Azure DNS Zone via Service Principal

### Introduction 
For the Azure provider to work, you will need to obtain a service principal within Azure, that has permissions to the DNS zone, within your resource group and subscription.

To create this, you will need the `azure-cli` and `jq` installed.

This guide assumes you already have an existing DNS zone created within Azure. If you don't, you can create one with the Azure CLI:

```
az group create --name "MyResourceGroupName" --location "uksouth"
az network dns zone create --resource-group "MyResourceGroupName" --name "myphonebookdomain.tld"
```
You may wish to substitute the location to a more suitable one nearer to you.

### Creating the service principal

For phonebook to be able to manage Azure DNS records, it requires access of `DNS Zone Contributor`, and Reader to the resource group containing the DNS zones themselves. More permissive levels will also work, but using the principle of least access is highly reccomended.

To create the service principle and grant permissions, you can run the below:

```bash
SP_NAME="MyPhoneBookServicePrincipal"
RG_NAME="MyResourceGroupName"
ZONE_NAME="myphonebookdomain.tld

SP=$(az ad sp create-for-rbac --name $SP_NAME)
SP_APP_ID=$(echo $SP | jq -r '.appId')
SP_APP_PW=$(echo $DNS_SP | jq -r '.password')

DNS_ID=$(az network dns zone show --name $ZONE_NAME --resource-group $RG_NAME --query "id" --output tsv)

az role assignment create --role "Reader" --assignee $SP_APP_ID --scope $DNS_ID
az role assignment create --role "Contributor" --assignee $SP_APP_ID --scope $DNS_ID

TENANT_ID=$(az account show --query tenantId -o tsv)
SUB_ID=$(az account show --query id -o tsv)

echo "AZURE_ZONE_NAME = $ZONE_NAME"
echo "AZURE_RESOURCE_GROUP = $RG_NAME"
echo "AZURE_SUBSCRIPTION_ID = $SUB_ID"
echo "AZURE_TENANT_ID = $TENANT_ID"
echo "AZURE_CLIENT_ID = $SP_APP_ID"
echo "AZURE_CLIENT_SECRET = $SP_APP_PW"
```

save the output of this in your preferred secure storage. You cannot retrieve the password post creation.

## Example DNSIntegration records

Create a DNSIntegration to start using your Azure zone with Phonebook

```yaml
apiVersion: se.quencer.io/v1alpha1
kind: DNSIntegration
metadata:
  name: azure
spec:
  provider:
    name: azure
  zones:
    - mydomain.com
  secretRef:
    name: azure-secrets
    keys:
      - name: "AZURE_ZONE_NAME"
        key: "zoneName"
      - name: "AZURE_RESOURCE_GROUP"
        key: "rgName"
      - name: "AZURE_SUBSCRIPTION_ID"
        key: "subId"
      - name: "AZURE_TENANT_ID"
        key: "tenantId"
      - name: "AZURE_CLIENT_ID"
        key: "clientId"
      - name: "AZURE_CLIENT_SECRET"
        key: "clientSecret"
```

If you wish to use environment variables over secrets:
```yaml
apiVersion: se.quencer.io/v1alpha1
kind: DNSIntegration
metadata:
  name: azure
spec:
  provider:
    name: azure
  zones:
    - mydomain.com
  env:
    - name: PHONEBOOK_PROVIDER
      value: azure
    - name: AZURE_ZONE_NAME
      value: zoneName
    - name: AZURE_RESOURCE_GROUP
      value: rgName
    - name: AZURE_SUBSCRIPTION_ID
      value: subId
    - name: AZURE_CLIENT_ID
      value: clientId
    - name: AZURE_CLIENT_SECRET
      value: clientSecret
    - name: AZURE_TENANT_ID
      value: tenantId
```

## Deploying

Now you can deploy with the normal command:
```
helm upgrade --install phonebook phonebook/phonebook \
  --namespace phonebook-system \
  --create-namespace \
  --values values.yaml
```
