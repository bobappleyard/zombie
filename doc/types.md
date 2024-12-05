Zombie is dynamically-typed. It has a small set of built-in types, along with a mechanism to create
new types. It also supports a simple type dispatch mechanism called _generic operations_.

## is?

    (is? x type)

Given an arbitrary object `x` and a type object `type`, test whether the object inhabits that type.

## make-type

    (make-type name cell-count)

Create a new type with a (string) name and a (number) cell-count.

## type-name

    (type-name type)

## make

    (make type)

Make a new object inhabiting the provided type. The cells will be void.

## cell-ref

    (cell-ref type x idx)
    (cell-set! type x idx value)

Retrieve and update cells in x give a type and an index.
