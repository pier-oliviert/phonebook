---
title: 'Errors'
date: 2024-09-20T18:29:19-04:00
draft: false
weight: 5
---

This is a list of all errors that are coded (PB#_NUM_: ...) in Phonebook. The library encodes all of its errors in this format to give more context to the user when an error comes up.

# General Error Codes

|PB#Number|Title|Description|
|:----|-|-|
|PB#0001|**Provider undefined**|Phonebook requires a valid provider to be defined through the `PHONEBOOK_PROVIDER` environment variable. The list of available provider is [available here]({{< ref "/providers" >}}).|
|PB#0002|**DNS Record not found**||
|PB#0003|**Provider could not delete the DNS record**||
|PB#0100|**Provider missing information**|Phonebook failed to initialized a client for the specified provider, more information can be found in the error message and in the provider's [section]({{< ref "/providers" >}}).|

# Provider Specific Error Codes

## Azure
|Number|Title|Description|
|:----|-|-|
|PB-AZ-#0001|Azure Client ID Not Found|Phonebook failed to find a valid client ID from a secret or env-var for the azure provider|
|PB-AZ-#0002|Azure Client Secret Not Found|Phonebook failed to find a valid client secret from a secret or env-var for the azure provider|
|PB-AZ-#0003|Azure Tenant ID Not Found|Phonebook failed to find a valid tenant ID from a secret or env-var for the azure provider|
|PB-AZ-#0004|Azure Subscription ID Not Found|Phonebook failed to find a valid subscription ID from a secret or env-var for the azure provider|
|PB-AZ-#0005|Azure Zone Name Not Found|Phonebook failed to find a valid zone name from a secret or env-var for the azure provider|
|PB-AZ-#0006|Azure Resource Group Not Found|Phonebook failed to find a valid resource group from a secret or env-var for the azure provider|
|PB-AZ-#0007|Unable to Create Azure Credential|Phonebook was unable to create an Azure credential using the provided information|
|PB-AZ-#0008|Unable to Create Azure DNS Client|Phonebook was unable to create an Azure DNS client using the provided information|
|PB-AZ-#0009|Failed to Create Resource Record Set|Phonebook failed to create a resource record set for Azure DNS|
|PB-AZ-#0010|Failed to Create Azure DNS Record|Phonebook failed to create an Azure DNS record|
|PB-AZ-#0011|Failed to Delete Azure DNS Record|Phonebook failed to delete an Azure DNS record|
|PB-AZ-#0012|CNAME Record Can Only Have One Target|Phonebook attempted to create a CNAME record with multiple targets, which is not allowed|
|PB-AZ-#0013|Invalid MX Record|Phonebook encountered an invalid MX record format|
|PB-AZ-#0014|Invalid SRV Record|Phonebook encountered an invalid SRV record format|
|PB-AZ-#0015|Unsupported Record Type|Phonebook encountered an unsupported DNS record type for Azure DNS|

## AWS
|Number|Title|Description|
|:----|-|-|
|PB-AWS-#0001|Failed to Load AWS Configuration|Phonebook failed to load the AWS configuration|
|PB-AWS-#0002|Zone ID Not Found|Phonebook failed to find a valid AWS Zone ID from a secret or env-var|
|PB-AWS-#0003|Failed to Create DNS Record|Phonebook failed to create a DNS record in AWS Route 53|
|PB-AWS-#0004|Failed to Delete DNS Record|Phonebook failed to delete a DNS record in AWS Route 53|
|PB-AWS-#0005|Unsupported Record Type|Phonebook encountered an unsupported DNS record type for AWS Route 53|

## Cloudflare
|Number|Title|Description|
|:----|-|-|
|PB-CF-#0001|API Key Not Found|Phonebook failed to find a valid Cloudflare API key from a secret or env-var|
|PB-CF-#0002|Zone ID Not Found|Phonebook failed to find a valid Cloudflare Zone ID from a secret or env-var|
|PB-CF-#0003|Unable to Create Cloudflare Client|Phonebook was unable to create a Cloudflare client using the provided information|
|PB-CF-#0004|Multiple Targets Not Supported|Phonebook attempted to create a DNS record with multiple targets, which is not supported by Cloudflare|