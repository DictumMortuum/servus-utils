DEPENDENCIES = ./cmd/model.go ./cmd/telnet.go ./cmd/util.go ./cmd/db.go
GOARGS = -trimpath -buildmode=pie -mod=readonly -modcacherw -ldflags="-s -w"

build: format
	go build $(GOARGS) -o srr  ./cmd/restart_router.go
	go build $(GOARGS) -o scsa ./cmd/connection_status_modem_a.go $(DEPENDENCIES)
	go build $(GOARGS) -o scsb ./cmd/connection_status_modem_b.go $(DEPENDENCIES)
	go build $(GOARGS) -o scsc ./cmd/connection_status_modem_c.go $(DEPENDENCIES)
	go build $(GOARGS) -o scsd ./cmd/connection_status_modem_d.go $(DEPENDENCIES)
	go build $(GOARGS) -o sds  ./cmd/download_subtitles.go
	go build $(GOARGS) -o scrape ./cmd/scrape.go $(DEPENDENCIES)

format:
	gofmt -s -w ./cmd ./pkg

clean:
	rm srr scsa scsb sds
