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

## Running a Server
The Inki server is available as a Docker container, you will need to setup your
server configuration file and mount it into the container to allow keys to be
published.

```yml
---
users:
  - name: root
    keyring: |
      -----BEGIN PGP PUBLIC KEY BLOCK-----
      Version: GnuPG v2
      ....
      ....
      ....
      -----END PGP PUBLIC KEY BLOCK-----
```

```sh
docker run --rm -p 3000:3000 -v "./config.yml:/etc/inki/server.yml" sierrasoftworks/inki:latest
```

Inki's server stores its configuration in memory, as its use case involves
providing transient key access to various servers. Stopping the container will
therefore remove any active keys and they will need to be added again.

## Adding a Key
Inki uses an HTTP API to add keys, requiring that a request to add a key is
sent as a signed PGP message with the JSON payload describing the key to be
added.

Due to the design, you can add keys using `curl` and the `gpg` command line
tools, alternatively Inki's command line can be used to submit the keys if
you find that easier.

### Using Inki
```sh
inki key add http://user@inki_server:3000 \
  --file ssh_key.pub \
  --pgp-key pgp_private_key.gpg \
  --expire 12h
```

### Using Curl
```sh
cat <<JSON
{
  "username": "user",
  "expire": "2016-12-25T00:00:00Z",
  "key": "$(cat ssh_key.pub)"
}
JSON | gpg --clearsign | curl -X POST http://inki_server:3000/api/v1/keys
```

## Using the Keys
Inki is designed to work with `sshd`'s AuthorizedKeysCommand to prevent situations
where a lack of disk space prevents you from accessing the server, as well as
avoiding corruption of your `authorized_keys` file. This has the added benefit
of allowing you to use Inki in conjunction with your existing set of `authorized_keys`.

To use Inki, you will need to create a script which calls the Inki agent to gather
the list of authorized keys.

```sh
#!/bin/bash
# $1 :  The username of the account that someone is attempting to sign in with

inki keys list http://$1@inki_server:3000 --authorized-keys

# You can also use this, if you don't want to have inki installed on your server
# curl http://inki_server:3000/api/v1/user/$1/authorized_keys
```

Then set the Inki agent as your AuthorizedKeysCommand in `/etc/ssh/sshd_config`

```
AuthorizedKeysCommand=/opt/my-inki-script
```