# Go macro

### A go tool to analyze your code and generate boilerplate in various formats

This module provides a tool taking Go code as input, and generating boilerplate code in the following formats :

- Go : JSON support for union types, SQL CRUD operations and utility data structure.
- SQL (Postgres) : creation statements and JSON validation functions
- TypeScript : type definitions and Axios API
- Dart : type definitions and JSON routines

## Module overview

### `analysis`

This package provides an extension of the standard `go/types` package with support for enums and union types. It knows nothing about the output targets.

### `generator`

This package uses the result provided by `analysis` to actually generate the code.

## Code directives and conventions

This module tries to be as smart and general as possible, but relies on special comments when
desambiguation is required.

- Definition of a constant which is not an enumeration: add `// gomacro:no-enum`
- Rely on an external generated file: add the `gomacro-extern:"<pkg>:<targetFile>"` tag to struct fields
- Types with name containing "Date" and with underlying time.Time are considered as date
