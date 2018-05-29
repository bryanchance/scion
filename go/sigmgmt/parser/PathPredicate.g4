grammar PathPredicate;

WHITESPACE: [ \r\n\t]+ -> skip;
DIGITS: '0' | [1-9] [0-9]*;
HEX_DIGITS: ('a' .. 'f' | 'A' .. 'F' | [0-9])+;
AS: HEX_DIGITS ':' HEX_DIGITS ':' HEX_DIGITS;

ANY: 'ANY' | 'any';
ALL: 'ALL' | 'all';
NOT: 'NOT' | 'not';

selector: DIGITS '-' (AS | DIGITS) '#' DIGITS;
condAny: ANY '(' cond (',' cond)* ')';
condAll: ALL '(' cond (',' cond)* ')';
condNot: NOT '(' cond ')';

cond: condAll | condAny | condNot | selector;
pathPredicate: cond EOF;
