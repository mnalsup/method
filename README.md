# Method

A CLI configured request client. Intended to be used in a similar manner as
Postman, however requests are stored in files and environment variables are
substituted.

## Installation
```bash
go get github.com/mnalsup/method
cd $GO_HOME/src/github.com/mnalsup/method

make install
```

## Usage
```bash
method <file>.yaml
```

Basic file structure:
```yaml
url: https://myurl:8080/graphql
method: POST 
headers:
  Content-Type: "application/json"
authenticationHook:
  triggers:
  - onJsonValue:
    path: "error.message"
    value: "Access Denied"
  - onHttpStatus:
    - 401
  requestPath: ~/method/auth/authReq.yaml
  jsonParseBodyPath: "token"
  bearerToken: {}
body:
  query: "query {
    Thing(Id: 123) {
      ThingID,
      Content
    }
  }"
```

For more see samples directory.

### Authentication Hook
Authentication hooks have triggers which are conditions that will cause the hook
to fire. Method will make the configured authentication request and use a json
path to parse the value required. This, for instance in the BearerToken
authType, will be placed in the Authorization header.

Trigger Options:

```yaml
- onHttpStatus:
  - 401
  - 403
  - etc...
```

```yaml
- onJsonValue
  path: "errors.0.message"
  value: "Access Denied"
```

The results of the authentication will be cached in a tmp file adjacent to the
original request. To view this cached request definition:
```bash
cat .<myrequest>.tmp.yaml
```

#### Authentication Strategies

Authentication strategies are an object within the authenticationHook that
method will read from.

A general formatted header might look like this, example is styled after the
bearer token format.
```
authenticationHook:
  authHeader:
    formatString: "Bearer %s"
    header: Authorization
```

An easy way to use a bearer token is just to use that strategy. Currently has
no options for modifications. Adds the token to the `Authorization` header after
Bearer.
```
authenticationHook:
  bearerToken: {}
```

Environment variable substitution is the simplest way. Just configure the hook
to set an environment variable and method will re-load the definition with the
environment variable set.
```
body:
  authToken: ${METHOD_MY_SERVICE_TOKEN}
authenticationHook:
  environmentVariable:
    variable: METHOD_MY_SERVICE_TOKEN
```


### Environment Substitution
Use environment variable replacements in your file by using the ${var} syntax
e.g:
```
url: ${METHOD_MY_QA_URL}/mypath
```


## Contributions

Contributions are welcome. Especially for use with new request and response
content types which I have been adding as I have need of them.
