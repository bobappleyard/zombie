# Zombie
Lisp-like language that targets WASM

## Implementation Plan

Bootstrap a compiler in Zombie, based on a very basic interpreter in Go. This can be found in the
`bootstrap` directory. The interpreter is very quick and dirty (the main evaluation logic was done
in an evening, the builtins and stdlib over the course of a weekend). Alongside this, we write a
very basic WASM runtime.

1. Using the interpreter, compile the compiler's dependencies.
2. Using the interpreter, compile the compiler (A).
3. Use wat2wasm (from WABT) to create executables.
4. Using the runtime + compiler (A), compile the compiler's dependencies.
5. Using the runtime + compiler (A), compile the compiler (B).
6. Using the runtime + compiler (B), compile the entire system.

Once we have a working system, this bootstrapping process will be less important.

This will hopefully maximise the reuse of the compiler/runtime/library code.
