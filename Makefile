.DEFAULT_GOAL := dev

ifeq ($(OS),Windows_NT)
HUGO=.bin/hugo.exe
else
HUGO=.bin/hugo
endif

download-sub:
	git submodule update --init

build-sub: cyberchef-prod

cyberchef-prod:
	npm install -g grunt-cli
	cd submodules/CyberChef && npm i && grunt prod
	cp -r submodules/CyberChef/build/prod static/CyberChef

download-hugo:

ifeq ($(OS),Windows_NT)
	wget -O .bin/hugo.zip https://github.com/gohugoio/hugo/releases/download/v0.85.0/hugo_extended_0.85.0_Windows-64bit.zip \
		&& 7z e .bin/hugo.zip -aoa -o.bin hugo.exe && rm .bin/hugo.zip
else
	wget -O- https://github.com/gohugoio/hugo/releases/download/v0.85.0/hugo_extended_0.85.0_Linux-64bit.tar.gz \
		| tar xzf - hugo && mv hugo .bin/
endif

dev:
	$(value HUGO) server -DF

build-dev:
	$(value HUGO) -DF

build:
	$(value HUGO)
