Token manager for OpenID Connect. Currently only supports direct
access (resource owner password credentials grant), which sends the
username and password to the authentication server.

# Authentication server setup

```
  took add <protocol> <options>
```

where:

 * protocol: oidc-direct-access


Options for oidc-direct-access:
 * -n Name of the configuration
 * -c Client ID
 * -s Client secret
 * -u Server URL, including domain, excluding protocol specific paths

```
  took add oidc-direct-access -n prod -c 12345 -s abcdef -u https://myserver/realms/myrealm
```

# Getting tokens

Once the authentication server is defined, use the following command to get a token:

```
  took token <name> [username]
```
where name is the configuration name used during add. If username is not given,
it'll be prompted. This will print out the active token if there is one,
renew if necessary, or ask for username/password to get a new token. For example:

```
  took token prod
  eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICI4MVRfd09aY29VekRDUXlhSnNYTXloUjhHQTlranViOEF6d1A3dTgzaDY4In0.eyJqdGkiOiJlY2UyM...
```


```
  took token -e prod
  Authentication: Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICI4MVRfd09aY29VekRDUXlhSnNYTXloUjhHQTlranViOEF6d1A3dTgzaDY4In0.eyJqdGkiOiJlY2UyM...
```

Use these additional flags to force renewal or re-authentication:


Renew token using the refresh token:
```
  took token -r <name>
```

Re-authenticate:
```
  took token -f <name>
```

Took can maintain tokens for multiple users. If username is ommitted, the last username will be used:

```
  took token prod user1
  <token for user1>
  took token prod
  <token for user1>
```

