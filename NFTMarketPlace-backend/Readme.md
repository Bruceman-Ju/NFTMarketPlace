
``` bash
docker build -t nft-marketplace-backend .
docker run -p 8080:8080 --env-file .env nft-marketplace-backend
```


``` bash
swag init -g cmd/main.go
```

``` bash
docker-compose -f docker-compose.yml up -d
```