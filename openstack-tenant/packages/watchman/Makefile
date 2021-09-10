PACKAGE = watchman
VERSION = 2021.08.23.00
SHA256SUM = a15bd0779725d2f2f511e50539ffd0b8a7a61ed793ca93f6300c10ef97b6091f

DISTRIBUTION = cirrus
COMPONENT = main

DISTPKG = watchman-v$(VERSION)-linux
PKGURL = https://github.com/facebook/watchman/releases/download/v$(VERSION)/$(DISTPKG).zip
DOWNLOAD = watchman.tar.gz

PKGDIR = $(PACKAGE)_$(VERSION)

all: setup_files
	../build.sh $(PACKAGE) $(VERSION) $(DISTRIBUTION) $(COMPONENT)

setup_files:
	mkdir $(PKGDIR)
	mkdir -p \
		$(PKGDIR)/usr/local/bin \
		$(PKGDIR)/usr/local/lib \
		$(PKGDIR)/etc/systemd/system
	wget -O $(DOWNLOAD) $(PKGURL)
	test `sha256sum $(DOWNLOAD) | awk '{print $$1}'` == $(SHA256SUM)
	unzip $(DOWNLOAD)
	cp $(DISTPKG)/bin/* $(PKGDIR)/usr/local/bin
	cp $(DISTPKG)/lib/* $(PKGDIR)/usr/local/lib
	cp watchman.service $(PKGDIR)/etc/systemd/system/watchman.service
	chmod a+x $(PKGDIR)/usr/local/bin/watchman{,ctl}

clean:
	$(RM) -r $(PKGDIR) $(DISTPKG)
	$(RM) $(PKGDIR).deb $(DOWNLOAD)
