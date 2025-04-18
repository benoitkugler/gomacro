# Go macro

### A go tool to analyze your code and generate boilerplate in various formats

This module provides a tool taking Go code as input and generating boilerplate code in the following languages :

- Go : JSON support for union types (`generator/go/unions`), SQL CRUD operations (`generator/go/sqlcrud`) and random data structure generation (`generator/go/randdata`).
- SQL (Postgres) : creation statements and JSON validation functions (`generator/sql`)
- TypeScript : type definitions and Axios API (`generator/typescript`)
- Dart : type definitions and JSON routines (`generator/dart`)

## CLI usage

### Single file mode

Provide a go source file as input, followed by the desired generation mode in the form
`<action>:<outputile>`. Several actions may be specified.

`./gomacro myinput.go sql:ouput.sql`

## Module overview

### `analysis`

This package provides an extension of the standard `go/types` package with support for enums, union types and time and date. It knows nothing about the output targets, but serves as a shared base.

Package `analysis/httpapi` provides a scanner to extract API urls and types. It only supports the Echo framework, but is modular, so that adding support for other frameworks should be quick.

Package `analysis/sql` adds a convertor from Go types to SQL ones and some logic about table relations.

### `generator`

This package uses the result provided by `analysis` to actually generate the code.

The `generator/typescript` package only targets the Axios Javascript library, but the
TypeScript types generator could be easily reused to support other methods.

## Code directives and conventions

This module tries to be as smart and general as possible, but relies on special comments when
desambiguation is required.

- The following struct fields are ignored:
  - unexported
  - with a 'json' tag '-'
  - with a 'gomacro' tag 'ignore'
- Definition of a constant which is not an enumeration: add `// gomacro:no-enum`
- Rely on an external generated file: add the `gomacro-extern:"<pkg>:<mode1>:<targetFile1>:<mode2>:<targetFile2>"` tag to struct fields
- Types with name containing "Date" and with underlying time.Time are considered as date
- Struct fields with `gomacro-data:"ignore"` are ignored from random data generation.
- SQL foreign keys are detected with types following the ID<table> convention or tagged with `gomacro-sql-foreign:"<table>"`, and an implicit constraint is generated. The `gomacro-sql-on-delete:"<action>"` may be provided to add for instance a cascade.
- Custom SQL constraints may be provided with struct comments of the form `// gomacro:SQL <constraint>`. The constraint is always prefixed by `ALTER TABLE <table>`. The Go struct names are replaced by their appropriate SQL equivalents. Enum value may be used inside comments with the syntax `#[<TypeName>.<EnumConstant>]`.
- SQL guard fields may be defined with the tag `gomacro-sql-guard:"<SQL value>"`
