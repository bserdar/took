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

Once the authentication server is defined, use

```
  took token <name>
```
where name is the configuration name used during add.

```
  took token prod
```

This would authenticate the user if necessary, retrieve the token, and display it.

```
  took token -e prod
```

This would print the output in the form:

```
Authentication: Bearer <token>
```


