Token manager command line for OpenID Connect. Currently supports
authorization flow and direct access (resource owner password
credentials grant).

# What does it do?

The main purpose of took is to maintain tokens for API
invocations. Once things are setup, you can run:

```
   took token myapi myuser
```
and it should either print out the token for myuser to call myapi,
or take you through authentication and then print the token. If the
token is expired and if there is a refresh token, it should get you
a new token without any further interaction. Once you have a valid
token, you can do:

```
   curl -H "Authorization: Bearer `took token myapi myuser`" http://myapi
```
or
```
   curl -H `took token -e myapi myuser`" http://myapi
```


# Authentication server setup

```
  took add <protocol> <options>
```

where:

 * protocol: oidc


Required options for oidc:
 * -n Name of the configuration. This is the 'myapi' parameter in the above examples
 * -c Client ID
 * -s Client secret
 * -u Server URL, including domain, excluding protocol specific paths
 * -b Callback URL


The following sets up a configuration called 'prod' using OIDC authorization flow:
```
  took add oidc -n prod -c 12345 -s abcdef -u https://myserver/realms/myrealm -b http://callback
```

Then, when you run
```
  took token prod myuser
```
It will ask you to visit a URL. That URL will authenticate the user, and redirect to the
callback URL, 'http://callback'. Copy this URL, and paste it to the command line, and it 
should print out a new token.

## Hack: Bypassing the server login page

It might be possible to describe the authentication form used by you server, so took can emulate
what the browser does to authenticate a user. When you go to the login page with the browser,
inspect the HTML page, and identify the forms and input fields. For instance, my server has the following
form:
```
<form id="kc-form-login" class="form-horizontal" onsubmit="login.disabled = true; return true;" 
action="https://sso.someserver/auth/realms/myrealm/login-actions/authenticate?code=QWI1Bmwm0&amp;execution=bca7381b-65b-4196-936c-7f8941f121&amp;client_id=security-admin-console&amp;tab_id=uRub-YYUVuk" method="post">
  <div class="form-group">
    <div class="col-xs-12 col-sm-12 col-md-4 col-lg-3">
       <label for="username" class="control-label">Username or email</label>
    </div>
    <div class="col-xs-12 col-sm-12 col-md-8 col-lg-9">
      <input tabindex="1" id="username" class="form-control" name="username" value="" type="text" autofocus autocomplete="off" />
     </div>
  </div>
  <div class="form-group">
    <div class="col-xs-12 col-sm-12 col-md-4 col-lg-3">
       <label for="password" class="control-label">Password</label>
    </div>
    <div class="col-xs-12 col-sm-12 col-md-8 col-lg-9">
      <input tabindex="2" id="password" class="form-control" name="password" type="password" autocomplete="off" />
   </div>

```
This HTML page has a form with id="kc-form-login", containing two input fields: username and password. 
You can define this structure with the -F flag:

```
took add oidc -n myapi -s 123 -b http://callback -c abc -u https://myserver \
 -F '{"id":"kc-form-login","usernameField":"username","passwordField":"password","fields":[{"input":"username","prompt":"User name"},\
    {"input":"password","prompt":"Password","password":true}]}'
```

When a new token is requested, took will ask for the username and password fields, submit the HTML
form, and get the tokens.


## Direct Access Grants Flow

Took supports direct access grants. In this flow, took asks username and password, and sends 
them to the authentication server.

```
  took add oidc -n prod-direct -c 12345 -s abcdef -u https://myserver/realms/myrealm -p
```
To use this, the authentication server must be configured to support 
direct access grants flow for this client.


# Multiple users 

Took can maintain tokens for multiple users. If username is ommitted, the last username will be used:

```
  took token prod user1
  <token for user1>
  took token prod
  <token for user1>
```

