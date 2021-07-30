.DEFAULT_GOAL := dev

HUGO=.bin/hugo
HUGO_VERSION=0.86.0

FUNCTION_SRC=functions/src/hello
FUNCTION_BIN=functions/_deploy/hello

$(FUNCTION_BIN) : $(FUNCTION_SRC)
	go build -o $@ github.com/eternal-flame-ad/yumechi.jp/$^

download-static:
	

download-sub:
	git submodule update --init

.bin/hugo:
	($(value HUGO) version | grep v$(value HUGO_VERSION)) || \
	(wget -O- https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_extended_${HUGO_VERSION}_Linux-64bit.tar.gz \
		| tar xzf - hugo && mv hugo .bin/)

download: download-static download-sub .bin/hugo

dev: download
	$(value HUGO) server -DF

build-dev: download
	$(value HUGO) -DF

build: download $(FUNCTION_BIN)
	rm -rf public/** || true
	$(value HUGO) --minify

clean:
	rm -rf resources/_gen/** || true
	rm -rf public/** || true

.PHONY: dev build-dev build download-sub download-static download .bin/hugo clean