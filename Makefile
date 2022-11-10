default: generate

build:
	mkdir -p build
	go build -o build/galleries src/galleries.go
	go build -o build/secure src/secure.go

galleries: build
	build/galleries --albums ~/sync/personal/jacoblewallen.com/content/albums

# This is only necessary to rebuild the music gallery area, because of how things work.
dynamic:
	cd dynamic && make
	cp dynamic/public/bundle.js jacoblewallen.com/static/js/dynamic.js
	cp dynamic/public/music.css jacoblewallen.com/static/css/

generate: galleries dynamic generate-hugo generate-zola

generate-hugo:
	rm -rf jacoblewallen.com/public
	cd jacoblewallen.com && hugo
	cd jacoblewallen.com && ../tools/postprocess.sh

generate-zola:
	rm -rf site/public
	cd site && zola build

upload:
	rsync -vua --delete jacoblewallen.com/public/ espial.me:live/public/

docker:
	cd docker && docker build -t jlewallen/personal .

clean:
	rm -rf node_modules

test:
	cd jacoblewallen.com && hugo server

.PHONY: docker clean build galleries dynamic music
