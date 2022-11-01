# Attendance System

[![codecov](https://codecov.io/gh/arcio-uk/attendance-system/branch/main/graph/badge.svg?token=0KE1EXR4ZO)](https://codecov.io/gh/arcio-uk/attendance-system)
[![CircleCI](https://dl.circleci.com/status-badge/img/gh/arcio-uk/attendance-system/tree/main.svg?style=svg&circle-token=686548e1f85c6054c50c276016bb1759be01dac8)](https://dl.circleci.com/status-badge/redirect/gh/arcio-uk/attendance-system/tree/main)

---

Attendance system API written in golang.

## Running

### Windows

```bash
go build; .\attendance-system.exe
```

### Linux

```bash
go build && ./attendance-system
```

### Docker

note: ensure bind address in .env is `0.0.0.0`

```bash
docker build -t attendance-system .
docker run -p 8010:8010 attendance-system
```

## Configuration

Create a .env file with the following environment variables:

| property          | meaning                                                          |
| ----------------- | ---------------------------------------------------------------- |
| `BIND_ADDR`       | The address to bind to, i.e: `0.0.0.0`                           |
| `BIND_PORT`       | The port to bind to i.e: `8010`                                  |
| `DB_URL`          | The database hostname or ip                                      |
| `DB_PORT`         | The port that the database runs on                               |
| `DB_NAME`         | The name of the attendance database                              |
| `DB_USERNAME`     | The username of the database                                     |
| `DB_PASSWORD`     | The password of the database                                     |
| `DB_MAX_CONNS`    | The maximum amount of open connections to have with the database |
| `JWT_ICAL_SCRET`  | The key for the HS512 signature of the ICAL jwt                  |
| `JWT_SECRET_FILE` | The JWT (ed512) public key file                                  |
| `SSL_MODE`        | enable/disable for SSL connection to DB                          |
| `NONCE_TOGGLE`    | true = enable nonces, false = disable                            |
| `XFORDWARD`       | true = log XFordwarded-For as ip address, false = use ip address |

The JWT public key must be stored in hex.

## Logging

Errors are logged to `stdout` and, audit logs to the database. To see errors you might want to
invoke the program with something such as

```bash
go build && bash -c "./attendance-system >> log"
```
