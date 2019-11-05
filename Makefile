test:
	go test github.com/textileio/grpc-ipfs-lite/server

ios:
	$(eval FLAGS := $$(shell govvv -flags | sed 's/main/github.com\/textileio\/grpc-ipfs-lite\/common/g'))
	env go111module=off gomobile bind -ldflags "-w $(FLAGS)" -v -target=ios github.com/textileio/grpc-ipfs-lite/mobile
	mkdir -p mobile/dist/ios/ && cp -r Mobile.framework mobile/dist/ios/
	rm -rf Mobile.framework

android:
	$(eval FLAGS := $$(shell govvv -flags | sed 's/main/github.com\/textileio\/grpc-ipfs-lite\/common/g'))
	env go111module=off gomobile bind -ldflags "-w $(FLAGS)" -v -target=android -o mobile.aar github.com/textileio/grpc-ipfs-lite/mobile
	mkdir -p mobile/dist/android/ && mv mobile.aar mobile/dist/android/