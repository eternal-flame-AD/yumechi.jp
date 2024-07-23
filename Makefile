.DEFAULT_GOAL := dev

HUGO=.bin/hugo
HUGO_VERSION=0.129.0
FUNCTIONS := functions/_deploy/hello

functions/_deploy/hello : functions/src/hello functions/src/hello/**
	cd functions/ && \
		go build -o $@ github.com/eternal-flame-ad/yumechi.jp/$(firstword $^)

functions: $(FUNCTIONS)

.bin/gimme: 
	wget -O.bin/gimme https://raw.githubusercontent.com/travis-ci/gimme/master/gimme
	chmod +x .bin/gimme

download-static: .bin/gimme


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

build: download functions
	rm -rf public/** || true
	if [ "${URL}" != "" ]; then \
		sed -i 's/yumechi.jp/$(URL:https://%=%)/g' config/_default/config.toml ; \
	fi
	$(value HUGO) --minify

clean:
	rm -rf functions/_deploy/** || true
	rm -rf resources/_gen/** || true
	rm -rf public/** || true

.PHONY: dev build-dev build download-sub download-static download .bin/hugo clean functions