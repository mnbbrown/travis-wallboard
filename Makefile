
.PHONY: client

build: client
	mkdir -p __build__
	go build -o __build__/travis-wallboard

client:
	pushd client && yarn build && popd
	statik -src=./client/build/
