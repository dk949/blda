include config.mk
all:

clean:

install: all
	mkdir -p ${DESTDIR}${PREFIX}/bin/
	install blda ${DESTDIR}${PREFIX}/bin/

uninstall:
	rm -f ${DESTDIR}${PREFIX}/bin/blda

.PHONY: all clean install uninstall
