default: docker

docker:
	cd docker && docker build -t jlewallen/personal .

build:
	mkdir -p build
	go build -o build/galleries src/galleries.go
	go build -o build/secure src/secure.go

clean:
	rm -rf node_modules

test:
	cd jacoblewallen.com && hugo server

galleries: build
	build/galleries --albums ~/dropbox/personal/jacoblewallen.com/content/albums

generate: galleries
	rm -rf jacoblewallen.com/public
	cd jacoblewallen.com && hugo
	cd jacoblewallen.com && ../tools/postprocess.sh

upload:
	rsync -vua --delete jacoblewallen.com/public/ espial.me:live/public/

.PHONY: docker clean build galleries
