include $(GOROOT)/src/Make.inc

TARG=librato
GOFILES=librato.go

include $(GOROOT)/src/Make.pkg

all: uninstall clean install
	make -C cmd/librato uninstall clean install

uninstall:
	rm -f $(GOROOT)/pkg/$(GOOS)_$(GOARCH)/$(TARG).a
	rm -f $(GOROOT)/pkg/$(GOOS)_$(GOARCH)/github.com/rcrowley/$(TARG).a
	rm -rf $(GOROOT)/src/pkg/github.com/rcrowley/$(TARG)
	make -C cmd/librato uninstall

.PHONY: all uninstall
