dev build
```
go build -ldflags "-X main.buildVersion=v1.0.0 -X main.buildDate=$(date -u +'%Y-%m-%dT%H:%M:%SZ') -X main.buildCommit=$(git rev-parse HEAD)" main.go
```

prod build

```
go build -ldflags "-X main.buildVersion=v1.0.0 -X main.buildDate=$(date -u +'%Y-%m-%dT%H:%M:%SZ') -X main.buildCommit=$(git rev-parse HEAD) -s -w" main.go
```