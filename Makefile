SHELL := /bin/bash
TARGETS = tabtokv

# http://docs.travis-ci.com/user/languages/go/#Default-Test-Script
test:
	go test -v

bench:
	go test -bench=.

imports:
	goimports -w .

fmt:
	go fmt ./...

vet:
	go vet ./...

all: fmt test
	go build

install:
	go install

clean:
	go clean
	rm -f coverage.out
	rm -f $(TARGETS)
	rm -f tabtokv-*.x86_64.rpm
	rm -f debian/tabtokv*.deb
	rm -rf debian/tabtokv/usr

cover:
	go get -d && go test -v	-coverprofile=coverage.out
	go tool cover -html=coverage.out

tabtokv:
	go build cmd/tabtokv/tabtokv.go

# ==== packaging

deb: $(TARGETS)
	mkdir -p debian/tabtokv/usr/sbin
	cp $(TARGETS) debian/tabtokv/usr/sbin
	cd debian && fakeroot dpkg-deb --build tabtokv .

REPOPATH = /usr/share/nginx/html/repo/CentOS/6/x86_64

publish: rpm
	cp tabtokv-*.rpm $(REPOPATH)
	createrepo $(REPOPATH)

rpm: $(TARGETS)
	mkdir -p $(HOME)/rpmbuild/{BUILD,SOURCES,SPECS,RPMS}
	cp ./packaging/tabtokv.spec $(HOME)/rpmbuild/SPECS
	cp $(TARGETS) $(HOME)/rpmbuild/BUILD
	./packaging/buildrpm.sh tabtokv
	cp $(HOME)/rpmbuild/RPMS/x86_64/tabtokv*.rpm .
