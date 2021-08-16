# Missing tests
connection-handling.go is not tested at all!

# General shortcomings
## Not a MUD yet!
It's just a telnet based server right now...

## Command parsing
The command parsing needs to be context sensitive! We need the player during the parse. Corollary: this means we can display a help page that is context sensitive!

## ANSI escape codes
- We should ask the telnet client what terminal capabilities it has
    - Done! we now check if "xterm" or "ansi" is present in the returned terminal type.
- We should have a "markup" language for output that allows for easy colorization (based on terminal capabilities) 
    - Done!
- We should have a consisten colorization of output. Maybe also have a simpler markup language. maybe
    - #fg_red, #bg_red, #fg_red_bright, etc