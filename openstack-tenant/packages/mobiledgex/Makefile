PACKAGE = mobiledgex
VERSION = 4.0.0-alpha4
DISTRIBUTION = stratus
COMPONENT = main

PKGDIR = $(PACKAGE)_$(VERSION)

all: builddir setup_files
	../build.sh $(PACKAGE) $(VERSION) $(DISTRIBUTION) $(COMPONENT)

builddir:
	mkdir $(PKGDIR)

$(PKGDIR)/etc/mobiledgex/% $(PKGDIR)/usr/local/bin/%: %
	mkdir -p $(dir $@)
	cp $< $@
	chmod a+x $@

$(PKGDIR)/etc/systemd/system/%: %
	mkdir -p $(dir $@)
	cp $< $@

setup_files: \
	$(PKGDIR)/etc/mobiledgex/install-k8s-master.sh \
	$(PKGDIR)/etc/mobiledgex/install-k8s-node.sh \
	$(PKGDIR)/etc/systemd/system/mobiledgex.service \
	$(PKGDIR)/usr/local/bin/mobiledgex-init.sh

clean:
	$(RM) -r $(PKGDIR)
	$(RM) $(PKGDIR).deb
