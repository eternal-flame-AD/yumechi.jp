.DEFAULT_GOAL := dev

HUGO=.bin/hugo
HUGO_VERSION=0.86.0

download-static:
	curl -fsSL https://cdn.jsdelivr.net/npm/mathjax@3/es5/tex-chtml.js  > static/js/mathjax.js

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

build: download
	$(value HUGO)

.PHONY: dev build-dev build download-sub download-static download .bin/hugo