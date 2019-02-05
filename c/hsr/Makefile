.PHONY: all clean install uninstall format tidy indent

CC = clang
CFLAGS ?= -Wall -Werror -O2 -DFORTIFY_SOURCE=2 -fstack-protector-all -fPIC -shared
VPP_LIBS = -lvnet -lvppinfra -lsvm -lvlib
LDFLAGS ?= -lpthread $(VPP_LIBS) -Wl,-Bstatic -Wl,-Bdynamic -Wl,--no-undefined
LIBNAME= scion_hsr.so

SOURCES=scion.c cli.c node.c
ifeq ($(D),1)
CFLAGS += -g -O0
endif

LIB_DIR = ../lib
INC = -I$(LIB_DIR)

TARGETS = $(LIBNAME)
#INSTALL = ../../bin/hsr
INSTALL = .
PREFIX = ..
all: $(TARGETS)

clean:
	rm -f *.d $(TARGETS)

# Compile .c files, while also automatically generating dependencies so they
# can be recompiled if any of the included header files changed.
-include *.d
$(LIBNAME): $(SOURCES)
	$(CC) $(CFLAGS) $(INC) -MMD -o $@ $? $(LDFLAGS)

install: $(INSTALL)

$(INSTALL): $(LIBNAME) 
	cp -f $(LIBNAME) $@

#  -nbad : no braces after declaration
#  -bap : braces after procedures
#  -br : braces on if line
#  -ce : else on previous closing brace line
#  -brs : braces on struct declaration
#  -l : max line length
#  -i : indent spaces
#  -ppi : indent on macros
#  -nbc : no new line after comma in a declaration
#  -bbo : break long lines before boolean operators
#  -hnl : honour already existing new lines in long lines
#  -sc : stars at the beginning of comment blocks
#  -ncdb : no comment delimiter on blank lines
#  -ci : line continuation indentation
#  -lp : aligned contiuation line to previous open parenthesis, brace, etc
#  -cli : case label indentation
#  -di : declaration indentation
#  -nfc1 : do not format first column comments
#  -nfca : do not format comments
#  -nprs : no space after parenthesis
#  -pcs : space after function call
#  -psl : type of procedure name before its name
#  -sai : space after if
#  -saw : space after while
#  -saf : space after for
#  -ncs : no space after cast
#  -sob : swallow optional blank lines
#  -ss : if semicolon on same line as loop statement, add space before it
#  -nut : no tabs
#
INDENT_OPTIONS= -nbad -bap -br -ce -brs -l100 -i4 -ppi3 -nbc -bbo -hnl -sc -ncdb \
				-ci8 -lp -cli0 -di1 -nfc1 -nfca -nprs -pcs -psl -sai -saw -saf \
				-ncs -sob -ss -nut -il1

indent:
	indent $(INDENT_OPTIONS) $(SOURCES)

tidy:
	clang-tidy $(SOURCES) -fix -checks="readability-braces-around-statements" -- $(CFLAGS)

uninstall:
	rm -f $(INSTALL)
