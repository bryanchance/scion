.PHONY: all clean install uninstall format tidy indent

# Use clang 6.0 instead of clang 3.8, as the latter complains with the following error:
#
# /usr/include/vlib/buffer_funcs.h:112:34: error: cannot convert between vector values of different size ('u64x2'
# (vector of 2 'u64' values) and 'int')
# u64x2_store_unaligned ((b1 << CLIB_LOG2_CACHE_LINE_BYTES) + off, b + 2);
#
CC = clang-6.0
TIDY= clang-tidy-6.0
CFLAGS ?= -Wall -Werror -O2 -DFORTIFY_SOURCE=2 -fstack-protector-all -fPIC -march=corei7
VPP_LIBS = -lvnet -lvppinfra -lsvm -lvlib -lcrypto
LDFLAGS ?= -lpthread $(VPP_LIBS) -Wl,-Bstatic -Wl,-Bdynamic -Wl,--no-undefined
LIBNAME= scion_hsr.so

SOURCES = scion.c bypass.c intf.c scion_input.c scion_format.c
OBJS = $(SOURCES:.c=.o)
INDENT_FILES = $(SOURCES) scion.h intf.h scion_packet.h scion_error.h scion_extensions.h

# XXX Enable debugging by default during development
D=1

ifeq ($(D),1)
CFLAGS += -gdwarf-4 -O0 -DCLIB_DEBUG
endif

LIB_DIR = ../lib
INC = -I$(LIB_DIR)

TARGETS = $(LIBNAME)
INSTALL = .
PREFIX = ..

all: $(TARGETS)

clean:
	rm -f *.d $(OBJS) $(TARGETS) *.e

# Compile .c files, while also automatically generating dependencies so they
# can be recompiled if any of the included header files changed.
-include *.d
%.o: %.c
	$(CC) $(CFLAGS) $(INC) -MMD -o $@ -c $<

$(LIBNAME): $(OBJS)
	$(CC) $(CFLAGS) $(INC) -shared -MMD -o $@ $^ $(LDFLAGS)

install: $(INSTALL)

%.e: %.c
	$(CC) $(CFLAGS) $(INC) -E -o $@ -c $<

.PHONY: cpp
# XXX pre-processor output files
cpp: $(SOURCES:.c=.e)

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
INDENT_OPTIONS= -nbad -bap -br -ce -brs -l100 -i4 -nbc -bbo -hnl -sc -ncdb \
				-ci8 -lp -cli0 -di1 -nfc1 -nfca -nprs -pcs -psl -sai -saw -saf \
				-ncs -sob -ss -nut -il1

indent:
	indent $(INDENT_OPTIONS) $(INDENT_FILES)
	indent $(INDENT_OPTIONS) $(INDENT_FILES)

tidy:
	$(TIDY) $(INDENT_FILES) -fix -checks="readability-braces-around-statements" -- $(CFLAGS)

uninstall:
	rm -f $(INSTALL)
