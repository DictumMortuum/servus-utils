build: format
	go build -trimpath -buildmode=pie -mod=readonly -modcacherw -ldflags="-s -w" -o srr ./cmd/restart_router.go
	go build -trimpath -buildmode=pie -mod=readonly -modcacherw -ldflags="-s -w" -o scs ./cmd/connection_status.go
	go build -trimpath -buildmode=pie -mod=readonly -modcacherw -ldflags="-s -w" -o sds ./cmd/download_subtitles.go

format:
	gofmt -s -w ./cmd
