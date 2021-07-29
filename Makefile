.DEFAULT_GOAL := dev

HUGO=.bin/hugo

download-static:
	curl -fsSL https://cdn.jsdelivr.net/npm/mathjax@3/es5/tex-chtml.js  > static/js/mathjax.js

download-sub:
	git submodule update --init

download-hugo:
	wget -O- https://github.com/gohugoio/hugo/releases/download/v0.85.0/hugo_extended_0.85.0_Linux-64bit.tar.gz \
		| tar xzf - hugo && mv hugo .bin/

dev:
	$(value HUGO) server -DF

build-dev:
	$(value HUGO) -DF

build:
	$(value HUGO)

.PHONY: dev build-dev build download-sub download-hugo download-static