PROJ_GO_DIR=github.com/oblique/create_ap

.PHONY: all
all:
	mkdir -p .gopath/src/$(dir $(PROJ_GO_DIR))
	[ -e .gopath/src/$(PROJ_GO_DIR) ] || \
		ln -sf $(CURDIR) .gopath/src/$(PROJ_GO_DIR)
	GOPATH=$(CURDIR)/.gopath go build $(PROJ_GO_DIR)

.PHONY: clean
clean:
	rm -rf create_ap .gopath

.PHONY: install
install:
	cp create_ap /usr/bin/create_ap
	[ ! -d /lib/systemd/system ] || cp contrib/create_ap.service /lib/systemd/system
	mkdir -p /usr/share/bash-completion/completions
	cp contrib/bash_completion /usr/share/bash-completion/completions/create_ap

.PHONY: uninstall
uninstall:
	rm -f /usr/bin/create_ap
	[ ! -f /lib/systemd/system/create_ap.service ] || rm -f /lib/systemd/system/create_ap.service
	rm -f /usr/share/bash-completion/completions/create_ap
