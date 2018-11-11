# URL Short
URL Shortener in Go

## Endpoints
`HTTP POST /shorten` takes a payload `{"custom":"","url":""}` and attempts to shorten the url. If `"custom"` is provided, the app will attempt to use that. If not, it will automatically generate a short url

`HTTP GET /links/{short}` will redirect to the correct original url 

`HTTP GET /links/{short}/stats` will provide the total hits for the link, as well as a histogram of the last 7 days of hits

## Running Unit Tests
Make sure postgres is installed locally.

To run the suite of unit tests simply do:
```
make test
```

## Running in docker
```
make docker-compose
```

This will build the docker image with the go binary, and the postgres container. The container is exposed on port 8080, and the db container can be reached at port 15432.

To issue commands:
```
curl -X POST localhost:8080/shorten -d '{"url":"http://www.longlonglongurltobeshortened.com"}'
curl -X POST localhost:8080/shorten -d '{"url":"http://www.abcdefg.com","custom":"myurl"}'

curl -X GET localhost:8080/shorten/myurl

curl -X GET localhost:8080/shorten/myurl/stats
```
