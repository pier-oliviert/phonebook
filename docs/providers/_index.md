---
title: 'Providers'
date: 2024-09-20T10:38:15-04:00
draft: false
---

This is the complete list of all DNS providers supported by Phonebook. Each provider have different requirements, so please read the section below that is associated with the provider you want to use.

While all values are written as environment variables name (eg. MY_SEKRET), each of those fields can be sourced from kubernetes secret and mounted as file. If you use secrets, the keys in the secret needs to be all caps, like an environment variable, as Phonebook expects the file to be of the same format as an environment variable's name.

