test:
	go test github.com/textileio/grpc-ipfs-lite/server

ios:
	#$(eval FLAGS := $$(shell govvv -flags | sed 's/main/github.com\/textileio\/grpc-ipfs-lite\/common/g'))
	go get golang.org/x/mobile/cmd/...
	go mod vendor
	mkdir -p ~/go/src/github.com/textileio
	rm -rf ~/go/src/github.com/textileio/grpc-ipfs-lite
	cp -r $(PWD) ~/go/src/github.com/textileio/
	export GO111MODULE=off
	gomobile bind -ldflags -v -target=ios github.com/textileio/grpc-ipfs-lite/mobile

android:
	#$(eval FLAGS := $$(shell govvv -flags | sed 's/main/github.com\/textileio\/grpc-ipfs-lite\/common/g'))
	go get golang.org/x/mobile/cmd/...
	go mod vendor
	mkdir -p ~/go/src/github.com/textileio
	rm -rf ~/go/src/github.com/textileio/grpc-ipfs-lite
	cp -r $(PWD) ~/go/src/github.com/textileio/
	export GO111MODULE=off
	gomobile bind -ldflags -v -target=android -o mobile.aar github.com/textileio/grpc-ipfs-lite/mobile
	