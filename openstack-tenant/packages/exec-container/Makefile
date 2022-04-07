PACKAGE = exec-container
VERSION = 0.1

DISTRIBUTION = cirrus
COMPONENT = main

PKGDIR = $(PACKAGE)_$(VERSION)

all: build
	../build.sh $(PACKAGE) $(VERSION) $(DISTRIBUTION) $(COMPONENT)

build:
	mkdir -p $(PKGDIR)/usr/local/bin
	GOOS=linux GOARCH=amd64 go build -o $(PKGDIR)/usr/local/bin/exec-container \
	     ../../../../edge-cloud/exec-container

clean:
	$(RM) -r $(PKGDIR)
	$(RM) $(PKGDIR).deb $(DOWNLOAD)
