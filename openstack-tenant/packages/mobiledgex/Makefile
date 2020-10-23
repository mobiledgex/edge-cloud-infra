PACKAGE = mobiledgex
VERSION = 4.1.1-pre1
DISTRIBUTION = cirrus
COMPONENT = main

PKGDIR = $(PACKAGE)_$(VERSION)

all: openstack vsphere

openstack:
	$(MAKE) package

vsphere:
	$(MAKE) package FLAVOR=vsphere VERSION=$(VERSION)-vsphere

package: builddir setup_files
	cp dependencies.common dependencies
ifdef FLAVOR
	cat dependencies.$(FLAVOR) >>dependencies
endif
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
	$(PKGDIR)/etc/mobiledgex/cleanup-vm.sh \
	$(PKGDIR)/etc/mobiledgex/get-flavor.sh \
	$(PKGDIR)/etc/mobiledgex/setup-chef.sh \
	$(PKGDIR)/etc/systemd/system/mobiledgex.service \
	$(PKGDIR)/usr/local/bin/mobiledgex-init.sh

clean:
	$(RM) -r $(PKGDIR) $(PKGDIR)-vsphere
	$(RM) $(PKGDIR).deb $(PKGDIR)-vsphere.deb dependencies
