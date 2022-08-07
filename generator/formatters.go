package generator

import (
	"log"
	"os/exec"
	"sync"
)

// utility wrappers around command line tools
// to format Go, Dart and TypeScript code.

// Formatters provides format commands for Go, Dart, TypeScript and SQL.
// The zero value is a ready to use cache.
type Formatters struct {
	lock                                       sync.Mutex
	hasGoFmt, hasDartFmt, hasTsFmt, hasPsqlFmt *bool
}

type Format uint8

const (
	NoFormat Format = iota
	Go
	Dart
	TypeScript
	Psql
)

// check if the goimports command is working
// and caches the result
func (fmts *Formatters) hasGo() bool {
	fmts.lock.Lock()
	defer fmts.lock.Unlock()

	if fmts.hasGoFmt == nil {
		err := exec.Command("which", "goimports").Run()
		if err != nil {
			log.Printf("No formatter for Go (%s)", err)
		}
		fmts.hasGoFmt = new(bool)
		*fmts.hasGoFmt = err == nil
	}
	return *fmts.hasGoFmt
}

// check if the dart command is working
// and caches the result
func (fmts *Formatters) hasDart() bool {
	fmts.lock.Lock()
	defer fmts.lock.Unlock()

	if fmts.hasDartFmt == nil {
		err := exec.Command("dart", "format", "--help").Run()
		if err != nil {
			log.Printf("No formatter for Dart (%s)", err)
		}
		fmts.hasDartFmt = new(bool)
		*fmts.hasDartFmt = err == nil
	}
	return *fmts.hasDartFmt
}

// check if the prettier command is working
// and caches the result
func (fmts *Formatters) hasTypescript() bool {
	fmts.lock.Lock()
	defer fmts.lock.Unlock()

	if fmts.hasTsFmt == nil {
		err := exec.Command("npx", "prettier", "-v").Run()
		if err != nil {
			log.Printf("No formatter for Typescript (%s)", err)
		}
		fmts.hasTsFmt = new(bool)
		*fmts.hasTsFmt = err == nil
	}
	return *fmts.hasTsFmt
}

// check if the pg_format command is working
// and caches the result
func (fmts *Formatters) hasPsql() bool {
	fmts.lock.Lock()
	defer fmts.lock.Unlock()

	if fmts.hasPsqlFmt == nil {
		err := exec.Command("pg_format", "-v").Run()
		if err != nil {
			log.Printf("No formatter for Psql (%s)", err)
		}
		fmts.hasPsqlFmt = new(bool)
		*fmts.hasPsqlFmt = err == nil
	}
	return *fmts.hasPsqlFmt
}

// FormatFile format `filename`, if a formatter for `format` is found.
// It returns an error if the command failed, not if no formatter is found.
func (fr *Formatters) FormatFile(format Format, filename string) error {
	switch format {
	case Go:
		if fr.hasGo() {
			return exec.Command("goimports", "-w", filename).Run()
		}
	case Dart:
		if fr.hasDart() {
			return exec.Command("dart", "format", filename).Run()
		}
	case TypeScript:
		if fr.hasTypescript() {
			return exec.Command("npx", "prettier", "--write", filename).Run()
		}
	case Psql:
		if fr.hasPsql() {
			return exec.Command("pg_format", "-i", filename).Run()
		}
	}
	return nil
}
