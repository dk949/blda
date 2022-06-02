include config.mk
all: bldr

bldr: bldr.go
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
