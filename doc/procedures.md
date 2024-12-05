# Procedure Objects

A procedure object has two fields: `(code, data)`

* `code` -- func table index
* `data` -- vector of captured variables

There is a runtime function `call-procedure`. This:

1. Checks that the first argument is a procedure
2. Prepends the captured variables on the stack
3. Jumps into the func table at the index


