# Darkstorm Backend

This is a purposefully "simple" application backend made specifically for _my_ apps. It's purpose is to collect minimal (only what's absolutely necessary) amounts of data while still fulfilling all my needs. I've found that other, off the shelf options such as Firebase are a bit heavy on the data collection. Plus I like to make things :P.

## DB Structure

### API Key

```json
{
  id: "API Key",
  appID: "appID",
  death: -1, // unix timestamp when the key is no longer valid. -1 means there is not expected expiration (that can change in the future)
  perm: {
    user: true, // create and login users
    log: true, // log users
    crash: true, // crash reports
    // further permissions can be added as needed
  }
}
```

### User

Users are stored per backend and not per app.

```json
{
  id: "UUID",
  username: "username",
  password: "hashed password",
  salt: "password salt",
  email: "email",
  passwordChange: 0, // unix timestamp of last password change
  perm: {
    appID: "user", // Optional. Apps should have a default permission level if thier appID is not in perm.
  }
}
```

### Crash Reports

#### Individual Report

```json
{
  count: 1, // We do not store duplicates. If a duplicate does occur
  platform: "android",
  error: "error",
  stack: "stacktrace"
}
```

#### Crashes

```json
{
  id: "UUID",
  error: "error",
  firstLine: "first line of error",
  individual: [
    // Individual Crash Reports
  ]
}
```

## Requests

### Standard Header

Any request might or might not need these headers. These values can be authenticated via the `ParseHeader` function.

```json
{
  X-API-Key: "{API Key}",
  Authorization: "Bearer {JWT Token}" // No built-in functions require a JWT Token, but may be required by specific implementations.
}
```

### Error Response

If an error status code is returned then the body will be as follows.

```json
{
  errorCode: "Error value for internal use",
  errorMsg: "User error message", //This message is meant to be displayed to the user. May be empty.
}
```

### Users

> TODO: Add the ability to create users and log-in through third-parties (such as Google).

All requsests pertaining to users requires the `X-API-Key` header and the key must have the `users` permission.

#### Create User

> TODO: Email user to confirm.
>
> TODO: Screen username for offensive words and phrases.

Request:

> POST: /user/create

```json
{
  username: "Username",
  password: "Password", // Password must be 
  email: "Email",
}
```

Return:

```json
{
  username: "Username",
  token: "JWT Token"
}
```

If returned status is 401, the errorCode will be one of the following:

* usernameTaken
  * Username is already taken
* usernameDisallowed
  * Username is not allowed (due to offensive words/phrases)
* password
  * Password is to short or too long.
* email
  * Email is already linked to an account
* disallowed
  * Username contains words/phases that are not allowed

#### Login

Request:

> POST: /user

```json
{
  username: "Username",
  password: "Password",
}
```

Return:

```json
{
  token: "JWT Token",
  timeout: 0, // login attempt timeout (in seconds). If non-zero, token will be empty.
}
```

### Crash Report

Crash reports require the `X-API-Key` header and the key must match the URL's appID and have the `crash` permission

Request:

> POST: /{appID}/crash

```json
{
  id: "UUID", // This is an ignored value, but it is highly recommended to include it to prevent reporting the same crash multiple times.
  platform: "android",
  error: "error",
  stack: "stacktrace"
}
```
