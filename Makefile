test:
	go test github.com/textileio/grpc-ipfs-lite/server

ios:
	$(eval FLAGS := $$(shell govvv -flags | sed 's/main/github.com\/textileio\/go-textile\/common/g'))
	go get golang.org/x/mobile/cmd/...
	go mod vendor
	rm -rf ~/go/src/github.com/textileio/grpc-ipfs-lite
	mkdir -p ~/go/src/github.com/textileio
	cp -R $(PWD) ~/go/src/github.com/textileio/grpc-ipfs-lite
	cd ~/go/src/github.com/textileio/grpc-ipfs-lite
	export GO111MODULE=off
	env go111module=off gomobile bind -ldflags "-w $(FLAGS)" -v -target=ios github.com/textileio/grpc-ipfs-lite/mobile

android:
	$(eval FLAGS := $$(shell govvv -flags | sed 's/main/github.com\/textileio\/go-textile\/common/g'))
	go get golang.org/x/mobile/cmd/...
	go mod vendor
	rm -rf ~/go/src/github.com/textileio/grpc-ipfs-lite
	mkdir -p ~/go/src/github.com/textileio
	cp -R $(PWD) ~/go/src/github.com/textileio/grpc-ipfs-lite
	~/go/src/github.com/textileio/grpc-ipfs-lite
	export GO111MODULE=off
	env go111module=off gomobile bind -ldflags "-w $(FLAGS)" -v -target=android -o mobile.aar github.com/textileio/grpc-ipfs-lite/mobile
	