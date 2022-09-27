# URL Shortener API

It is a URL shortening service, made with Go + Fiber + Redis and used docker for building it.

## Used technologies

* [fiber](https://github.com/gofiber/fiber)
* [redis](https://github.com/go-redis/redis)
* [docker](https://www.docker.com/)
* [docker-compose](https://docs.docker.com/compose/)

## How to run

* Install [golang](https://go.dev/dl/) and [docker](https://www.docker.com/)


* File `.env` fill the variables


* Run `docker-compose up -d`


* App exposed on PORT 3000


* Redis Server exposed on PORT 6379


## API Endpoints

### GET `/:url`

Redirects the shortcode to original long URL.

Response payload

In case short code exists it responds with 301 redirect.

If the short code is expired or deleted, it responds like so:

HTTP Status Code: 404

```json
{
  "error": "short not found in database"
}
```

### POST `/api/urls`

Creates a new short code for given URL.

Request example

* url: string, required, http, https only 
* custom_short: string
* expiry: integer (hour), default = 24

```json
{
  "url": "https://stackoverflow.com/?newreg=c889c1d419ce40829cdb636ef52dd03a",
  "short": "stackoverflow",
  "expiry": "24"
}
```

Responses example

1. Case of successful creation

HTTP Status Code: 200

```json
{
  "url": "https://stackoverflow.com/?newreg=c889c1d419ce40829cdb636ef52dd03a",
  "short": "localhost:3000/stackoverflow",
  "expiry": 24,
  "rate_limit": 9,
  "rate_limit_reset": 30
}
```

2. Invalid request data

HTTP Status Code: 400

```json
{
  "error": "cannot parse JSON"
}
```

3. Invalid URL

HTTP Status Code: 400

```json
{
  "error": "invalid url"
}
```

4. Rate limit exceeded (a client can receive no more than 10 requests in 30 minutes)

HTTP Status Code: 503

```json
{
  "error":            "Rate limit exceeded",
  "rate_limit_reset": 29
}
```

5. Create an already existing url

HTTP Status Code: 403

```json
{
  "error": "URL custom short is already in use"
}
```
