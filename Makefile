default: generate

build:
	mkdir -p build
	go build -o build/galleries src/galleries.go
	go build -o build/secure src/secure.go

galleries: build
	build/galleries --albums ~/sync/personal/site-hugo/content/albums

# This is only necessary to rebuild the music gallery area, because of how things work.
dynamic:
	cd dynamic && make
	cp dynamic/public/bundle.js site-hugo/static/js/dynamic.js
	cp dynamic/public/music.css site-hugo/static/css/

generate: galleries dynamic generate-hugo generate-zola

generate-hugo:
	rm -rf site-hugo/public
	cd site-hugo && hugo
	cd site-hugo && ../tools/postprocess.sh

generate-zola:
	rm -rf site-zola/public
	cd site-zola && zola build

upload:
	rsync -vua --delete site-hugo/public/ espial.me:live/public/

docker:
	cd docker && docker build -t jlewallen/personal .

clean:
	rm -rf node_modules

test-hugo:
	cd site-hugo && hugo server

test-zola:
	cd site-zola && zola serve

.PHONY: docker clean build galleries dynamic music
