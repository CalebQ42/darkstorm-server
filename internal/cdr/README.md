# cdr-backend

Stupid backend for [CDR]("https://github.com/CalebQ42/CustomDiceRoller").

## APIs

### Dice

Dice sharing

> POST: /upload?key={api_key}

Upload a die.

Request Body:

```json
{
    // Die data
}
```

Note: Only allows up to 1MB of data. If over 1MB returns 413. Further limits might be imposed in the future.

Response:

```json
{
    "id": "die ID",
    "expiration": 0 // Unix time (Seconds) of expiration
}
```

> GET: /die/{die id}?key={api_key}

Get an uploaded die.

Response:

```json
{
    // die data minus uid
}
```
