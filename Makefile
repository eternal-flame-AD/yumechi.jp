.DEFAULT_GOAL := dev

HUGO=.bin/hugo
HUGO_VERSION=0.111.3

FUNCTIONS := functions/_deploy/hello functions/_deploy/comment_submit

functions/_deploy/hello : functions/src/hello functions/src/hello/**
	go build -o $@ github.com/eternal-flame-ad/yumechi.jp/$(firstword $^)

functions/_deploy/comment_submit : functions/src/comment_submit functions/src/comment_submit/**
	go generate ./...
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
	@if [ "${URL}" = "" ]; then \
		cp config.toml _build_config.toml; \
	else \
		sed 's/yumechi.jp/$(URL:https://%=%)/g' config.toml > _build_config.toml; \
	fi
	$(value HUGO) --config _build_config.toml --minify

clean:
	rm -rf functions/_deploy/** || true
	rm -rf resources/_gen/** || true
	rm -rf public/** || true

.PHONY: dev build-dev build download-sub download-static download .bin/hugo clean functions