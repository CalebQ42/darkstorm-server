# swassistant-backend

Custom backend for [SWAssistant](https://github.com/CalebQ42/SWAssistant). Extension of [darkstorm-backend](https://github.com/CalebQ42/darkstorm-server/tree/main/internal/backend)

## APIs

For `POST` requests, the `X-API-Key` http header must be set.

### Profiles

#### Upload profile to share

Character, vehicles, and minion profiles.

> POST: /profile?type={character|vehicle|minion}

Upload a profile. `type` query is required.

Request Body:

```json
{
  // profile data
}
```

Note: Only allows up to 5MB of data. If over 5MB returns 413. Further limits might be imposed in the future.

Response:

```json
{
  "id": "profile ID",
  "expiration": 0 // Unix time (Seconds) of expiration
}
```

#### Get a shared profile

> GET: /profile/{profileID}

Get an uploaded profile.

Response:

```json
{
  "type": "character|vehicle|minion",
  // profile data minus uid
}
```

### Rooms

All room requests must include both `X-API-Key` and `Authorization` headers.

#### Room list

> GET: /rooms

Get a list of rooms your currently a part of.

Response:

```json
[
  {
    "id": "room ID",
    "name": "room name",
    "owner": "username"
  }
]
```

#### Create new room

> POST: /rooms/new?name={roomName}

Create a new room. `name` query is required.

Response:

```json
{
  "id": "room ID",
  "name": "room name"
}
```

#### Get room info

> GET: /rooms/{roomID}

Get info about a room.

```json
{
  "id": "room ID",
  "name": "room name",
  "owner": "username",
  "users": [
    "username"
  ],
  "profiles": [
    "profile uuids"
  ]
}
```
