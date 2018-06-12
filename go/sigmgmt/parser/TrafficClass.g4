grammar TrafficClass;

WHITESPACE: [ \r\n\t]+ -> skip;
DIGITS: '0' | [1-9] [0-9]*;
HEX_DIGITS: ('a' .. 'f' | 'A' .. 'F' | [0-9])+;
NET: DIGITS '.' DIGITS '.' DIGITS '.' DIGITS '/' DIGITS;

ANY: 'ANY' | 'any';
ALL: 'ALL' | 'all';
NOT: 'NOT' | 'not';
SRC: 'SRC' | 'src';
DST: 'DST' | 'dst';
DSCP: 'DSCP' | 'dscp';
TOS: 'TOS' | 'tos';

matchSrc: SRC '=' NET;
matchDst: DST '=' NET;
matchDSCP: DSCP '=0x' (HEX_DIGITS | DIGITS);
matchTOS: TOS '=0x' (HEX_DIGITS | DIGITS);

condCls: 'cls=' DIGITS;
condAny: ANY '(' cond (',' cond)* ')';
condAll: ALL '(' cond (',' cond)* ')';
condNot: NOT '(' cond ')';

condIPv4: matchSrc | matchDst | matchDSCP | matchTOS;
cond: condAll | condAny | condNot | condIPv4 | condCls;
trafficClass: cond EOF;
