
compile:
	mkdir -p dist
	docker build -t mdnsforwarder:latest .
	docker run --rm -v $(CURDIR)/dist:$(shell docker run --rm -t mdnsforwarder:latest pwd)/dist -t mdnsforwarder:latest make compile-direct

pack:
	docker run --rm -v $(CURDIR)/dist:$(shell docker run --rm -t mdnsforwarder:latest pwd)/dist -t mdnsforwarder:latest make pack-direct

compile-pack: compile pack

compile-direct:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/mdnsforwarder-amd64 cli/mdnsforwarder.go
	GOOS=linux GOARCH=mips GOMIPS=softfloat go build -trimpath -ldflags="-s -w" -o dist/mdnsforwarder-mips cli/mdnsforwarder.go
	GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -trimpath -ldflags="-s -w" -o dist/mdnsforwarder-mipsle cli/mdnsforwarder.go
	GOOS=linux GOARCH=arm go build -trimpath -ldflags="-s -w" -o dist/mdnsforwarder-arm cli/mdnsforwarder.go
	GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/mdnsforwarder-arm64 cli/mdnsforwarder.go

pack-direct:
	upx -9 dist/mdnsforwarder-amd64
	upx -9 dist/mdnsforwarder-mips
	upx -9 dist/mdnsforwarder-mipsle
	upx -9 dist/mdnsforwarder-arm
	upx -9 dist/mdnsforwarder-arm64


build-docker:
	docker build -t cbrand/mdnsforwarder:latest -f Dockerfile.run .

publish-docker: build-docker
	docker push cbrand/mdnsforwarder:latest
