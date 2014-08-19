# Features
# ========
#
# Default behavior is to run all tests (including integration tests, see more below),
# build all binaries and build a .deb-package.
# All files included in "./deb" will be included in the .deb-package. "./deb/etc/project.conf"
# will be installed at "/etc/project.conf" when installing the .deb-package.
#
# deb: Default behavior.
#
# bin: Build all binaries.
#
# test: Run all tests.
#
# major: Increment major version by 1. Modifies ./deb/DEBIAN/control.
#
# minor: Increment minor version by 1. Modifies ./deb/DEBIAN/control.
#
# patch: Increment patch version by 1. Modifies ./deb/DEBIAN/control.
#
# clean: Remove all build artifacts produced produced by make.
#
# goget: Runs "go get" for the project. 
#
SHELL := /bin/bash

go_package = prov/...

package_name = $(shell grep Package: deb/DEBIAN/control | cut -d ' ' -f 2)
package_version = $(shell grep Version: deb/DEBIAN/control | cut -d ' ' -f 2)

package_version_major = $(shell echo $(package_version) | cut -d '.' -f 1)
package_version_major_next = $(shell expr 1 + $(package_version_major))

package_version_minor = $(shell echo $(package_version) | cut -d '.' -f 2)
package_version_minor_next = $(shell expr 1 + $(package_version_minor))

package_version_patch = $(shell echo $(package_version) | cut -d '.' -f 3)
package_version_patch_next = $(shell expr 1 + $(package_version_patch))

package_full = $(package_name)-$(package_version)
package_file = $(package_full).deb

export GOPATH=$(CURDIR)

deb: test $(package_file)
.PHONY: deb

$(package_file): bin $(shell find deb -type f)
	mkdir -p $(package_full)/usr/bin
	cp -f bin/* $(package_full)/usr/bin/
	cp -rf deb/* $(package_full)/
	dpkg-deb --build $(package_full)
	rm -rf $(package_full)

bin: $(shell find $(CURDIR) -name '*.go')
	mkdir -p bin
	GOBIN=$(CURDIR)/bin go install $(go_package)

test:
	go test $(go_package)
.PHONY: test

major:
	sed -i 's/Version:.*/Version: $(package_version_major_next).$(package_version_minor).$(package_version_patch)/' deb/DEBIAN/control
.PHONY: major

minor:
	sed -i 's/Version:.*/Version: $(package_version_major).$(package_version_minor_next).$(package_version_patch)/' deb/DEBIAN/control
.PHONY: minor

patch:
	sed -i 's/Version:.*/Version: $(package_version_major).$(package_version_minor).$(package_version_patch_next)/' deb/DEBIAN/control
.PHONY: patch

clean:
	rm -rf bin pkg *.deb $(package_full)
.PHONY: clean

goget:
	go get ...
.PHONY: goget
