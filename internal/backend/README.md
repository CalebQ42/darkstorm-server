# Darkstorm Backend

This is a purposefully "simple" application backend made specifically for _my_ apps. It's purpose is to collect minimal (only what's absolutely necessary) amounts of data while still fulfilling all my needs. I've found that other, off the shelf options such as Firebase are a bit heavy on the data collection. Plus I like to make things :P.

## DB Structure

### API Key

```json
{
  id: "API Key",
  appID: "appID",
  death: -1, // unix timestamp (seconds) when the key is no longer valid. -1 means there is not expected expiration (that can change in the future)
  perm: {
    user: true, // create and login users
    count: true, // count users
    crash: true, // crash reports
    management: false, // managing
    // further permissions can be added as needed
  }
}
```

Optionally you can set a special AppID to be a management key. Setting a management key enables management requests.

### Count log

```json
{
  id: "UUID",
  platform: "android",
  Date: 20240519 // YYYYMMDD as int
}
```

### User

Users are stored per backend and not per app.

```json
{
  id: "uuid",
  username: "username",
  password: "hashed password",
  salt: "password salt",
  email: "email",
  fails: 0, // number of failed attemps in a row.
  timeout: 0, // unix timestamp (seconds) when current timeout ends.
  passwordChange: 0, // unix timestamp (seconds) of last password change
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
  appVersion: "v1.0.0", // Application version
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

`errorCode`'s returned from the main library:

* misconfigured
  * Backend is configured incorrectly (such as App returning nil crash table, but key has crash permission)
* invalidKey
  * API Key is invalid or does not have the needed permission for the request.
* invalidBody
  * Body of the request is malformed.
* unauthorized
  * User is not authorized for the given task or no user token is given.
* badRequest
  * Some part of your request is invalid
* internal
  * Server-side issue.

### Count

API Key must have the `count` permission.

Request:

> POST: /count

```json
{
  id: "uuid", // Should be an empty string on first request. If invalid or too old, a new UUID will be returned.
  platform: "web"
}
```

Returns:

```json
{
  id: "uuid"
}
```

### User Count

Get a count of users.

API Key must have the `management` permission.

`platform` query is optional (defaults to all).

Request:

> GET: /count?platform=all

With management key:

> GET: /{appID}/count?platform=all

Returns:

```json
{
  count: 0
}
```

### Users

> TODO: Add the ability to create users and log-in through third-parties (such as Google).

All requsests pertaining to users requires the `X-API-Key` header and the key must have the `users` permission.

Enabled by using `Backend.AddUserAuth`.

#### Create User

> TODO: Email user to confirm.
>
> TODO: Screen username for offensive words and phrases.

Request:

> POST: /user/create

```json
{
  username: "Username",
  password: "Password", // Allowed length: 12-128
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

* taken
  * Username or email is already taken
* usernameDisallowed
  * Username is not allowed (due to offensive words/phrases)
* password
  * Password is to short or too long.

#### Delete User

Requires either the `management` permission or a management key.

Request:

> DELETE: /user/{userID}

#### Login

Request:

> POST: /user/login

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
  error: "Error",
  timeout: 0, // login attempt timeout remaining (in seconds). If non-zero, token will be empty.
}
```

`token` and `error` are mutually exclusive.

Possible `error` values:

* timeout
  * Account is currently timed-out. The `timeout` value will be non-zero.
* invalid
  * Either the username or password is incorrect

#### Change Password

Request:

> POST: /user/changepassword

```json
{
  token: "JWT Token",
  old: "Old Password",
  new: "New Password"
}
```

### Crash Report

#### Report

API Key must have the `crash` permission.

Request:

> POST: /crash

Request Body:

```json
{
  id: "UUID", // This is an ignored value, but it is highly recommended to include it to prevent reporting the same crash multiple times.
  platform: "android",
  appVersion: "v1.0.0",
  error: "error",
  stack: "stacktrace"
}
```

#### Delete

API Key must have the `management` permission.

Request:

> DELETE: /crash/{crashID}

With management key:

> DELETE: /{appID}/crash/{crashID}

#### Archive

Archive an error, preventing error with these values to be ignored in the future. API Key must have the `management` permission.

Request:

> POST: /crash/archive

With management key:

> POST: /{appID}/crash/archive

Request Body:

```json
{
  error: "error",
  stack: "full stacktrace", // Archives will only match against a perfect match.
  platform: "all", // Limit the archive to a specific platform, or use "all".
}
```
