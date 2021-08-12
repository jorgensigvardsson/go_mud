# Missing tests

input-queue_test.go is incomplete - fix it!

connection-handling.go is not tested at all!

# General shortcomings
## Guarantee \r\n line endings
TELNET requires \r\n for line endings. Currently we're using fmt.Println() et al to format newlines, and on Linux they produce \n, not \r\n