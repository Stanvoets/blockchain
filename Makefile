PACKAGES_NOSIMULATION=$(shell go list ./... | grep -v '/simulation')
PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
CAT := $(if $(filter $(OS),Windows_NT),type,cat)
LEDGER_ENABLED ?= false
GOTOOLS = \
	github.com/golang/dep/cmd/dep \
	github.com/alecthomas/gometalinter \
	github.com/rakyll/statik
GOBIN ?= $(GOPATH)/bin

# process build tags

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq ($(WITH_CLEVELDB),yes)
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))s

BUILD_FLAGS := -tags "$(build_tags)"

all: install

########################################
### Build/Install

build:
ifeq ($(OS),Windows_NT)
	go build $(BUILD_FLAGS) -o build/bcnad.exe ./cmd/bcnad
	go build $(BUILD_FLAGS) -o build/bcnacli.exe ./cmd/bcnacli
else
	go build $(BUILD_FLAGS) -o build/bcnad ./cmd/bcnad
	go build $(BUILD_FLAGS) -o build/bcnacli ./cmd/bcnacli
endif

build-linux:
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

install:
	go install $(BUILD_FLAGS) ./cmd/bcnad
	go install $(BUILD_FLAGS) ./cmd/bcnacli
	@statik -src=cmd/bcnacli/lcd/swagger-ui -dest=cmd/bcnacli/lcd -f
	$(call go_install,rakyll,statik,v0.1.5)