include $(GOROOT)/src/Make.inc

TARG=github.com/rcrowley/go-librato
GOFILES=\
	collated.go\
	librato.go\
	simple.go\

include $(GOROOT)/src/Make.pkg

all: uninstall clean install
	make -C cmd/librato uninstall clean install

uninstall:
	rm -f $(GOROOT)/pkg/$(GOOS)_$(GOARCH)/$(TARG).a
	rm -rf $(GOROOT)/src/pkg/$(TARG)
	make -C cmd/librato uninstall

.PHONY: all uninstall
