# Description

There are two services:
- **gateway** - gateway service as proxy for other services;
- **api** - api service that implements all logic for api, including auth and weather parts;  

Swagger is available on [http://localhost:8080/docs/](http://localhost:8080/docs/). 

# Commands

Before running tests build the project:
```
go mod vendor
docker-compose up -d
```

To run tests:

```
make test
```

# CLI

To build jcli:
```
make build-cli
```

To create admin user:
```
jcli --rpcaddr localhost:9090 users createadmin --email <email> --password <password>
``` 

# To do

## Improvements

- Provide config not via parameters but as via env variables
- Add production/development logging with env
- Add weather service support with message queue
- Separate ACL from user model not to extract it each time we need the user. And search/remove will be faster
