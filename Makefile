default: docker

docker:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o shortener .
	docker build -t shortener .

docker-compose: docker
	docker-compose stop
	docker-compose rm -f
	docker-compose up -d
	
test:
	dropdb urlshort_test
	createdb urlshort_test
	psql -U postgres -d urlshort_test -a -f database/migration/*
	go test -v ./...