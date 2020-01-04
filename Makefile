default: docker

docker:
	cd docker && docker build -t jlewallen/personal .

build:
	mkdir -p build
	go build -o build/albums *.go

clean:
	rm -rf node_modules

test:
	cd jacoblewallen.com && hugo server

galleries: build
	build/albums --albums ~/dropbox/personal/jacoblewallen.com/content/albums

.PHONY: docker clean build galleries
