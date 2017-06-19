SHELL := $(shell env which bash)

all: build

include go.mk

# the name of the program being built
PROG := $(PKG_DIR_GO)/$(IMPORT_PATH).a

# configure and load the csi protobuf generator
CSI_PROTO_DIR := csi
include csi.mk

gocsi: $(PROG)
$(PROG): 	$(CSI_GOSRC) \
			$(filter-out %_test.go,$(wildcard *.go)) | $(GODEPS) $(PKG_DIR_GO)
	go install -pkgdir $(abspath $(PKG_DIR_GO))
	@echo $@

build: $(PROG)
	$(MAKE) -C mock build

MOCK_PLUGIN := mock/mock-csi-plugin.so
$(MOCK_PLUGIN):
	$(MAKE) -C mock $(@F)

test: $(PROG) $(MOCK_PLUGIN)
	CSI_PLUGINS="$(abspath $(MOCK_PLUGIN))" \
	go test -v -pkgdir $(abspath $(PKG_DIR_GO))

clean:
	go clean -i
	$(MAKE) -C mock clean

clobber: clean
	rm -fr $(CSI_PROTO_DIR) $(BUILD_DIR)
	$(MAKE) -C mock clobber

.PHONY: clean clobber test
