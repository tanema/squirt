Squirt
======

## Run it
go run ./cmd/squirt

## Decisions
- use lua type tables/buckets with different ways to interact with it.
- for num loop for general loop with classic form
- for in for key values
- self is always scoped in function defined on table, no special syntax, points to table
- default to local, no way to export to global other than returning
- assignment should be consistent
  - if targets and values match, its 1:1
  - if targets are less than values, the last value gets a bucket with the remaining values
  - if targets are more than values, the remaining are left null
  - this should work for spreads as well.
- Almost everything is a class except classes and func. So "string" is an Instance of String

## Milestone 3
- Refinements and autocontructors
  - [x] initial value
  - [x] `const`
  - [x] extracted and set refinements
  - [x] `type` error assigned an incorrect type
  - [x]  autocontructor
  - [x] `required` argument error in constructor if not passed
  - [x] `get`, `set` bool to set readonly or writeonly, func to allow custom setter/getter
- [x] indexing
  - [x] index out of range handling
  - [x] range indexing `source[1:4]`
- [x] `and` and `or` should return values other than bool. (`false || x` should return the value of x)
- [x] ternary, maybe solve by having if statements return last values. (`x = if true then a else b end`)
  - already have lua style ternary `condval and iftrue or iffalse`
- [x] minimized operator interfaces
  - [x] != should just be !(==)
  - [x] OpNot should just use tobool and then !
  - [x] gt, gte, lt, lte to just a single <=> compare type operator `__compare`
- [ ] require extension
  - [x] require stdlib
  - [x] caching
  - [ ] directory require
  - [ ] require paths LOAD_PATHS
  - [ ] package management solution
- [ ] test tooling
  - [ ] cli test flag
  - [ ] assert tooling
  - [ ] describe, test description blocks

## Milestone 4
- [ ] localization. Configuration at the package level of what to translate each keyword into so that packages that use different languages still interop
- [ ] stdlib
  - io
  - file
  - os
  - http

## Type annotations
- [ ] Parse time type annotation checking
  - [ ] type annotations with inference
  - [ ] function parameter matching
  - [ ] function call parameter type and count check
  - [ ] assignment type mismatch (maybe only on class attributes)

## Education
- [ ] Docs!
- [ ] animate parsing with excerpts
- [ ] animate runtime with excerpts

## Implement it
- [x] function call
- [x] assignment
- [x] table def
- [x] table index
- [x] Binary operators
- [x] unary operators
- [x] block scope statements
  - [x] if statements (no break, return propagates)
  - [x] function def (no break, return)
  - [x] do block (no break, return propagates)
  - [x] for num loop (break, return propagates)
  - [x] for in loop (break, return propagates)
  - [x] while loop (break, return propagates)
  - [x] until loop (break, return propagates)
  - [x] break
  - [x] next
  - [x] return values
- [x] self
- [x] function call parameters
  - [x] named spread as last parameter

## Change it
- [x] long strings start and stop with \`
- [x] change comments from -- to //
  - [x] long comments with /* */
- [x] default to local
- [x] change function to func or fn
- [x] take out `:` and always scope `self`
- [x] replace `~=` with `!=`
- [x] use ! unary instead of `not`
- [x] replace `=` in table construction with `:`
- [x] Name
- [x] take out optional () in function calls
- [x] take out goto/label
  - [x] add in `next`

## Second Milestone implementation
- [x] runtime error with stacktrace and pointer to problem
- [x] squirt naming
- [x] parser error snippets with clear pointer to problem
- [x] vim syntax highlighting
- [x] string interpolation
- [x] table ops `<<`, `+`, `-`
- [x] `toString()` and `toNumber()`
- [x] spread after table to unpack
  - [x] assignment
  - [x] function call values
  - [x] table constructor
  - [x] return values
- [x] multiple return values
  - [x] assignment
  - [x] function call values
  - [x] table constructor
- [x] `delete()` method for removing from tables
- [x] string indexing
- [x] `require`
- [x] `eval`
- [x] `typeof()`
- [x] incr, decr `-=`, `+=`, `--`, `++`

## Object Orientation
- [x] Inheritance
  - [x] [proposals](inheritance_proposals.md)
  - [x] parsing
  - [x] class definition
  - [x] class function definition
  - [x] method definition
  - [x] constructor
  - [x] super
  - [x] operator overriding
    - [x] unary
    - [x] binary
    - [x] delegate ops to classes for string, number, bool, table
      - [x] bool
      - [x] func
      - [x] nil
      - [x] number
      - [x] string
      - [x] table
    - [x] system
      - [x] call
      - [x] del
      - [x] tostring
      - [x] tobool
      - [x] index
      - [x] assignindex
    - [x] input args need to be classes
    - [x] Table keys need to be classes and they are not right now
    - [x] String index should be iindexable runtime.go:617
  - [x] inherit from std classes
- [x] error handling
  - [x] block level catch/rescue with `clean`, (func, do, for, while, repeat)
  - [x] throw errors with `spill`, will throw with string or class.
  - [x] @ protect unary, first assign value will be an error or nil.
