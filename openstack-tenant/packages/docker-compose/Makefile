PACKAGE = docker-compose
VERSION = 1.24.1
SHA256SUM = cfb3439956216b1248308141f7193776fcf4b9c9b49cbbe2fb07885678e2bb8a

DISTRIBUTION = cirrus
COMPONENT = main

PKGURL = https://github.com/docker/compose/releases/download/$(VERSION)/docker-compose-Linux-x86_64
DOWNLOAD = docker-compose

PKGDIR = $(PACKAGE)_$(VERSION)

all: setup_files
	../build.sh $(PACKAGE) $(VERSION) $(DISTRIBUTION) $(COMPONENT)

setup_files:
	mkdir $(PKGDIR)
	mkdir -p $(PKGDIR)/usr/local/bin
	wget -O $(DOWNLOAD) $(PKGURL)
	test `sha256sum $(DOWNLOAD) | awk '{print $$1}'` == $(SHA256SUM)
	mv $(DOWNLOAD) $(PKGDIR)/usr/local/bin
	chmod a+x $(PKGDIR)/usr/local/bin/$(DOWNLOAD)

clean:
	$(RM) -r $(PKGDIR)
	$(RM) $(PKGDIR).deb $(DOWNLOAD)
