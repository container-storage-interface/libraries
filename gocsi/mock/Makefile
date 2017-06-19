SHELL := $(shell env which bash)

all: build

# load the go build logic from the parent directory
include ../go.mk

# the name of the program being built
PROG := mock

# the program to build
SERVER_BIN := $(PROG)-csi-server

# adjust the binary names for windows
ifeq (windows,$(GOOS))
SERVER_BIN := $(SERVER_BIN).exe
else
# the name of the plug-in binary to build
PLUGIN_BIN := $(PROG)-csi-plugin.so
endif

# configure the csi protobuf generator
CSI_PROTO_DIR := csi

# load the csi protobuf generator from the parent directory
include ../csi.mk

SERVER_BIN_TMP := $(BIN_DIR_GO)/$(PROG)
$(SERVER_BIN_TMP): main.go $(CSI_GOSRC) | $(GODEPS) $(BIN_DIR_GO)
	GOBIN=$(abspath $(@D)) go install
$(SERVER_BIN): $(SERVER_BIN_TMP)
	cp -f $< $@

PLUGIN_BIN_TMP := $(PKG_DIR_GO)/$(IMPORT_PATH).a
$(PLUGIN_BIN_TMP): main.go $(CSI_GOSRC) | $(GODEPS) $(PKG_DIR_GO)
	go install -buildmode plugin -pkgdir $(abspath $(PKG_DIR_GO))
$(PLUGIN_BIN): $(PLUGIN_BIN_TMP)
	cp -f $< $@

build: $(SERVER_BIN)
# add the plug-in target to all if building on linux for linux
ifeq (linux_linux,$(GOOS)_$(GOHOSTOS))
build: $(PLUGIN_BIN)
endif

clean:
	rm -f $(SERVER_BIN) $(SERVER_BIN_TMP) $(PLUGIN_BIN) $(PLUGIN_BIN_TMP)

clobber: clean
	rm -fr $(CSI_PROTO_DIR) $(BUILD_DIR)

.PHONY: clean clobber
