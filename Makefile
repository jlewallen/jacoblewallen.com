default: generate

build:
	mkdir -p build
	go build -o build/galleries src/galleries.go
	go build -o build/secure src/secure.go

galleries: build
	build/galleries --albums ~/sync/personal/site-zola/content/albums

generate: generate-zola

generate-hugo:
	rm -rf site-hugo/public
	cd site-hugo && hugo
	cd site-hugo && ../tools/postprocess.sh

generate-zola:
	rm -rf site-zola/public
	cd site-zola && zola build

upload:
	rsync -vua --delete site-zola/public/ espial.me:live/public/

docker:
	cd docker && docker build -t jlewallen/personal .

clean:
	rm -rf node_modules site-zola/public

test:
	cd site-zola && zola serve

.PHONY: docker clean build galleries dynamic music
