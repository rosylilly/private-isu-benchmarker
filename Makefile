versions:=$(wildcard v*)

define MAKE_VERSION
	@echo "===> ${2} ${1}"
	@cd ${1} && make ${2} || true

endef

.PHONY: build
build:
	$(foreach version,$(versions),$(call MAKE_VERSION,${version},build))

.PHONY: run
run:
	$(foreach version,$(versions),$(call MAKE_VERSION,${version},run))

.PHONY: clean
clean:
	$(foreach version,$(versions),$(call MAKE_VERSION,${version},clean))
