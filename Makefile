PREFIX ?= /usr
BUILD_TAGS ?= nogui
VERSION ?= 1.0.1
PKG_NAME ?= ibus-bamboo-viet

.PHONY: all test test-c build vicore config-tool gui deb clean install fuzz fuzz-stress

all: build vicore config-tool gui

test:
	go test -tags $(BUILD_TAGS) -v ./...
	$(MAKE) test-c

fuzz:
	@mkdir -p bin
	go run ./cmd/fuzzer --mode=telex --count=500

fuzz-stress:
	@mkdir -p bin
	go run ./cmd/fuzzer --mode=telex --stress=true

fuzz-sentence:
	@mkdir -p bin
	go run ./cmd/fuzzer --sentence="$(SENTENCE)"


vicore:
	@mkdir -p bin
	go build -buildmode=c-shared -o bin/libvicore.so ./libvicore/

test-c: vicore
	@mkdir -p bin
	gcc -o bin/c_test tests/c_test.c -Lbin -lvicore -Wl,-rpath,'$$ORIGIN'
	./bin/c_test

config-tool:
	@mkdir -p bin
	go build -o bin/bamboo-viet-config ./cmd/config_tool/

gui:
	@mkdir -p bin
	@cp cmd/gui/bamboo_viet_gui.py bin/bamboo-viet-gui
	@chmod +x bin/bamboo-viet-gui

build:
	@mkdir -p bin
	go build -tags $(BUILD_TAGS) -o bin/ibus-engine-bamboo .

fuzzer:
	@mkdir -p bin
	go build -o bin/bamboo-viet-fuzzer ./cmd/fuzzer/

deb: build vicore config-tool gui fuzzer
	@echo "Creating Debian package staging directory..."
	@rm -rf packaging/staging
	@mkdir -p packaging/staging/DEBIAN
	@mkdir -p packaging/staging/usr/lib/ibus-bamboo
	@mkdir -p packaging/staging/usr/bin
	@mkdir -p packaging/staging/usr/lib
	@mkdir -p packaging/staging/usr/include
	@mkdir -p packaging/staging/usr/share/ibus/component
	@mkdir -p packaging/staging/usr/share/applications
	@mkdir -p packaging/staging/usr/share/metainfo
	@mkdir -p packaging/staging/usr/share/ibus-bamboo/data
	@mkdir -p packaging/staging/usr/share/ibus-bamboo/icons
	@cp packaging/debian/DEBIAN/* packaging/staging/DEBIAN/
	@cp bin/ibus-engine-bamboo packaging/staging/usr/lib/ibus-bamboo/
	@cp bin/bamboo-viet-config packaging/staging/usr/bin/
	@cp bin/bamboo-viet-gui packaging/staging/usr/bin/
	@cp bin/bamboo-viet-fuzzer packaging/staging/usr/bin/
	@cp cmd/gui/bamboo_viet_gui.py packaging/staging/usr/share/ibus-bamboo/
	@chmod 755 packaging/staging/usr/bin/bamboo-viet-config
	@chmod 755 packaging/staging/usr/bin/bamboo-viet-gui
	@chmod 755 packaging/staging/usr/bin/bamboo-viet-fuzzer
	@cp bin/libvicore.so packaging/staging/usr/lib/
	@cp bin/libvicore.h packaging/staging/usr/include/
	@cp data/bamboo.xml packaging/staging/usr/share/ibus/component/
	@cp data/ibus-setup-Bamboo.desktop packaging/staging/usr/share/applications/
	@cp data/org.bamboo_viet.control_panel.metainfo.xml packaging/staging/usr/share/metainfo/
	@cp -r data/* packaging/staging/usr/share/ibus-bamboo/data/
	@cp -r icons/* packaging/staging/usr/share/ibus-bamboo/icons/
	@dpkg-deb -b packaging/staging bin/$(PKG_NAME)_$(VERSION)_amd64.deb
	@echo "✓ Debian package created at bin/$(PKG_NAME)_$(VERSION)_amd64.deb"

clean:
	rm -rf bin/ packaging/staging/
