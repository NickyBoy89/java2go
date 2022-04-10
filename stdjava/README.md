# stdjava is a Go implementation of some of the Java-isms used when parsing the codebase

Currently, this includes:
* A generic `Ternary` function that takes in a condition, and outputs one of the two results
* Unsigned right shift (`>>>=` and `>>>`), which does right shifts, but fills the top bits with zeroes, instead of being sign-dependent
* Java's string `hashCode` function
* The `Optional<T>` type
