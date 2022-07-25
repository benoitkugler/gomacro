# Go macro

### A go tool to analyze your code and generate boilerplate in various formats

This module provides a tool taking Go code as input and generating boilerplate code in the following formats :

- Go : JSON support for union types (`generator/go/unions`), SQL CRUD operations and utility data structure.
- SQL (Postgres) : creation statements and JSON validation functions
- TypeScript : type definitions and Axios API (`generator/typescript`)
- Dart : type definitions and JSON routines (`generator/dart`)

## Module overview

### `analysis`

This package provides an extension of the standard `go/types` package with support for enums, union types and time and date. It knows nothing about the output targets, but serves as a shared base.

Package `analysis/httpapi` provides a scanner to extract API urls and types. It only supports the Echo framework, but is modular, so that adding support for other framework should be quick.

### `generator`

This package uses the result provided by `analysis` to actually generate the code.

The `generator/typescript` package only targets the Axios Javascript library, but the
TypeScript types generator could be easily reused to support other methods.

## Code directives and conventions

This module tries to be as smart and general as possible, but relies on special comments when
desambiguation is required.

- Definition of a constant which is not an enumeration: add `// gomacro:no-enum`
- Rely on an external generated file: add the `gomacro-extern:"<pkg>:<mode1>:<targetFile1>:<mode2>:<targetFile2>"` tag to struct fields
- Types with name containing "Date" and with underlying time.Time are considered as date
