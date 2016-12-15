# Inki
**Secure SSH key distribution with support for custom workflow logic**

Inki is a tool which makes it trivially easy to manage a dynamic list of SSH keys
on a host. This is achieved through a daemon which holds an in-memory list of
keys and provides an HTTP API via which new keys may be added, as well as a client
which consumes the API.

To prevent the possibility of bad actors registering keys against your hosts,
it is possible to configure Inki to require SSH keys to be PGP signed before they
are accepted.

## Features
 - **Support for multiple users**, allowing you to register keys for individual user
   accounts and potentially requiring unique PGP keys for individual users.
 - **Integrates with AuthorizedKeysCommand** to remove the need for modifications to
   your `authorized_keys` file and also enable Inki to add keys even when the host
   has no diskspace remaining.
 - **Straightforward HTTP API** to enable other services to quickly and easily integrate
   with it. You can even send commands using Curl if need be!

## Example
```
$ inki key add http://bpannell@inki.sierrasoftworks.com -f my_key.pub -p sign.key
Enter PGP key password:
Added keys:
 - Username:     bpannell
   Fingerprint:  7646dd89cbbcecbfeda2ba1d80ec9451
   Expires:      2016-12-15 14:30:42.9195054 +0000 UTC
  
$ inki key list http://bpannell@inki.sierrasoftworks.com
Authorized keys:
 - Username:     bpannell
   Fingerprint:  7646dd89cbbcecbfeda2ba1d80ec9451
   Expires:      2016-12-15 14:30:42.9195054 +0000 UTC

$ inki authorized-keys bpannell
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDArmZ5fyEt1V9KiGFuiZ...
```

## Use Case
Inki was originally designed to enable automated tools to request access to servers
for remediation purposes, allowing the servers to decide whether to allow the tool
access on a case-by-case basis and ensuring that credentials could be flexibly rotated
at any time.

That being said, it offers a great way to enable access to your servers using a PGP
key like your Keybase one and any SSH key, potentially saving you from the loss of
an SSH key while keeping your systems secure.