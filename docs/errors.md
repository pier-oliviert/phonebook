---
title: 'Errors'
date: 2024-09-20T18:29:19-04:00
draft: false
weight: 5
---

This is a list of all errors that are coded (PB#_NUM_: ...) in Phonebook. The library encodes all of its errors in this format to give more context to the user when an error comes up.

|PB#Number|Title|Description|
|:----|-|-|
|PB#0001|**Provider undefined**|Phonebook requires a valid provider to be defined through the `PHONEBOOK_PROVIDER` environment variable. The list of available provider is [available here](./providers.md).|
|PB#0002|**DNS Record not found**||
|PB#0003|**Provider could not delete the DNS record**||
|PB#0100|**Provider missing information**|Phonebook failed to initialized a client for the specified provider, more information can be found in the error message and in the provider's [section](./providers.md).|

