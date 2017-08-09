SHELL := $(shell env which bash)

# the name of the program being built
PROG := $(notdir $(shell pwd))

all: $(PROG)

# configure and load the csi protobuf generator
CSI_PROTO_DIR := csi
include csi.mk

# adjust the binary names for windows
ifeq (windows,$(GOOS))
PROG := $(PROG).exe
endif

$(PROG): main.go $(CSI_GOSRC)
	go install -v
	go build -v -o $@

clean:
	go clean -i && rm -fr $(PROG)

clobber: clean
	rm -f $(PROG)

.PHONY: clean clobber
