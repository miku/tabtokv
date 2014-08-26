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

publish: rpm-compatible
	cp tabtokv-*.rpm $(REPOPATH)
	createrepo $(REPOPATH)

rpm: $(TARGETS)
	mkdir -p $(HOME)/rpmbuild/{BUILD,SOURCES,SPECS,RPMS}
	cp ./packaging/tabtokv.spec $(HOME)/rpmbuild/SPECS
	cp $(TARGETS) $(HOME)/rpmbuild/BUILD
	./packaging/buildrpm.sh tabtokv
	cp $(HOME)/rpmbuild/RPMS/x86_64/tabtokv*.rpm .

# ==== vm-based packaging

PORT = 2222
SSHCMD = ssh -o StrictHostKeyChecking=no -i vagrant.key vagrant@127.0.0.1 -p $(PORT)
SCPCMD = scp -o port=$(PORT) -o StrictHostKeyChecking=no -i vagrant.key

# Helper to build RPM on a RHEL6 VM, to link against glibc 2.12
vagrant.key:
	curl -sL "https://raw.githubusercontent.com/mitchellh/vagrant/master/keys/vagrant" > vagrant.key
	chmod 0600 vagrant.key

# Don't forget to vagrant up :) - and add your public key to the guests authorized_keys
setup: vagrant.key
	$(SSHCMD) "sudo yum install -y sudo yum install http://ftp.riken.jp/Linux/fedora/epel/6/i386/epel-release-6-8.noarch.rpm"
	$(SSHCMD) "sudo yum install -y golang git rpm-build"
	$(SSHCMD) "mkdir -p /home/vagrant/src/github.com/miku"
	$(SSHCMD) "cd /home/vagrant/src/github.com/miku && git clone https://github.com/miku/tabtokv.git"

rpm-compatible: vagrant.key
	$(SSHCMD) "cd /home/vagrant/src/github.com/miku/tabtokv && GOPATH=/home/vagrant go get ./..."
	$(SSHCMD) "cd /home/vagrant/src/github.com/miku/tabtokv && git pull origin master && pwd && GOPATH=/home/vagrant make rpm"
	$(SCPCMD) vagrant@127.0.0.1:/home/vagrant/src/github.com/miku/tabtokv/*rpm .

