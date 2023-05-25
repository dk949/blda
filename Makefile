include config.mk
all: bldr


bldr: bldr.go
	sed 's/^const VERSION =.*$$/const VERSION = "$(VERSION)"/' bldr.go -i
	go build -o bldr bldr.go
	strip bldr


clean:
	rm -f bldr


install: all
	mkdir -p ${DESTDIR}${PREFIX}/bin/
	install bldr ${DESTDIR}${PREFIX}/bin/

uninstall:
	rm -f ${DESTDIR}${PREFIX}/bin/bldr

.PHONY: all clean install uninstall
