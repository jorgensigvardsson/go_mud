# Missing tests
connection-handling.go is not tested at all!

# General shortcomings
## Not a MUD yet!
It's just a telnet based server right now...

## Command parsing
The command parsing needs to be context sensitive! We need the player during the parse. Corollary: this means we can display a help page that is context sensitive!
- Command parsing is now context sensitive - the player is taken into consideration
- We now present a help page/command

## VT Escape codes
How about a dedicated prompt area?