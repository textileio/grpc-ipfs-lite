test:
	go test ./...

ios:
	go get golang.org/x/mobile/cmd/...
	go mod vendor
	rm -rf ~/go/src/github.com/textileio/grpc-ipfs-lite
	mkdir -p ~/go/src/github.com/textileio
	cp -R $(PWD) ~/go/src/github.com/textileio/grpc-ipfs-lite
	export GO111MODULE=off
	env go111module=off gomobile bind -ldflags "-w $(FLAGS)" -v -target=ios github.com/textileio/grpc-ipfs-lite/mobile

android:
	go get golang.org/x/mobile/cmd/...
	go mod vendor
	rm -rf ~/go/src/github.com/textileio/grpc-ipfs-lite
	mkdir -p ~/go/src/github.com/textileio
	cp -R $(PWD) ~/go/src/github.com/textileio/grpc-ipfs-lite
	export GO111MODULE=off
	env go111module=off gomobile bind -ldflags "-w $(FLAGS)" -v -target=android -o mobile.aar github.com/textileio/grpc-ipfs-lite/mobile
	