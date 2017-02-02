CC=gcc
CFLAGS=-Wall -Wextra -pedantic -std=c99 -O0
TESTFILE=known_byte_sequences.c

all: test64

test64:
	$(CC) $(CFLAGS) $(TESTFILE) -o test

clean:
	rm test
