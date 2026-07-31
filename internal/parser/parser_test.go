package parser

import (
	"path/filepath"
	"testing"

	"github.com/pgulb/aimap/internal/symbol"
)

func TestParseGoFile(t *testing.T) {
	path := filepath.Join("testdata", "src", "hello.go")
	syms, err := ParseGoFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(syms) == 0 {
		t.Fatal("expected at least one symbol")
	}

	byName := make(map[string]symbol.Symbol)
	for _, s := range syms {
		byName[s.Name] = s
	}

	greeting, ok := byName["Greeting"]
	if !ok {
		t.Fatal("expected symbol Greeting")
	}
	if greeting.Kind != symbol.Function {
		t.Errorf("Greeting.Kind = %v, want Function", greeting.Kind)
	}
	if greeting.LineStart < 1 {
		t.Errorf("Greeting.LineStart = %d, want > 0", greeting.LineStart)
	}

	config, ok := byName["Config"]
	if !ok {
		t.Fatal("expected symbol Config")
	}
	if config.Kind != symbol.Struct {
		t.Errorf("Config.Kind = %v, want Struct", config.Kind)
	}

	appVer, ok := byName["AppVersion"]
	if !ok {
		t.Fatal("expected symbol AppVersion")
	}
	if appVer.Kind != symbol.Constant {
		t.Errorf("AppVersion.Kind = %v, want Constant", appVer.Kind)
	}

	formatter, ok := byName["Formatter"]
	if !ok {
		t.Fatal("expected symbol Formatter")
	}
	if formatter.Kind != symbol.Interface {
		t.Errorf("Formatter.Kind = %v, want Interface", formatter.Kind)
	}

	getName, ok := byName["GetName"]
	if !ok {
		t.Fatal("expected symbol GetName")
	}
	if getName.Kind != symbol.Method {
		t.Errorf("GetName.Kind = %v, want Method", getName.Kind)
	}
	if getName.Parent != "Config" {
		t.Errorf("GetName.Parent = %q, want Config", getName.Parent)
	}
}

func TestParseGoGenericType(t *testing.T) {
	path := filepath.Join("testdata", "src", "hello.go")
	syms, err := ParseGoFile(path)
	if err != nil {
		t.Fatal(err)
	}

	byName := make(map[string]symbol.Symbol)
	for _, s := range syms {
		byName[s.Name] = s
	}

	list, ok := byName["List"]
	if !ok {
		t.Fatal("expected symbol List")
	}
	if list.Kind != symbol.Struct {
		t.Errorf("List.Kind = %v, want Struct", list.Kind)
	}
	if list.TypeParams != "[T any]" {
		t.Errorf("List.TypeParams = %q, want [T any]", list.TypeParams)
	}

	pair, ok := byName["Pair"]
	if !ok {
		t.Fatal("expected symbol Pair")
	}
	if pair.Kind != symbol.Struct {
		t.Errorf("Pair.Kind = %v, want Struct", pair.Kind)
	}
	if pair.TypeParams != "[T, U any]" {
		t.Errorf("Pair.TypeParams = %q, want [T, U any]", pair.TypeParams)
	}
}

func TestParseGoGenericMethod(t *testing.T) {
	path := filepath.Join("testdata", "src", "hello.go")
	syms, err := ParseGoFile(path)
	if err != nil {
		t.Fatal(err)
	}

	byName := make(map[string]symbol.Symbol)
	for _, s := range syms {
		byName[s.Name] = s
	}

	push, ok := byName["Push"]
	if !ok {
		t.Fatal("expected symbol Push")
	}
	if push.Kind != symbol.Method {
		t.Errorf("Push.Kind = %v, want Method", push.Kind)
	}
	if push.Parent != "List[T]" {
		t.Errorf("Push.Parent = %q, want List[T]", push.Parent)
	}
}

func TestParseGoGenericFunction(t *testing.T) {
	path := filepath.Join("testdata", "src", "hello.go")
	syms, err := ParseGoFile(path)
	if err != nil {
		t.Fatal(err)
	}

	byName := make(map[string]symbol.Symbol)
	for _, s := range syms {
		byName[s.Name] = s
	}

	newList, ok := byName["NewList"]
	if !ok {
		t.Fatal("expected symbol NewList")
	}
	if newList.Kind != symbol.Function {
		t.Errorf("NewList.Kind = %v, want Function", newList.Kind)
	}
	if newList.TypeParams != "[T any]" {
		t.Errorf("NewList.TypeParams = %q, want [T any]", newList.TypeParams)
	}
}

func TestParseGoInterfaceMethods(t *testing.T) {
	path := filepath.Join("testdata", "src", "hello.go")
	syms, err := ParseGoFile(path)
	if err != nil {
		t.Fatal(err)
	}

	byName := make(map[string]symbol.Symbol)
	for _, s := range syms {
		byName[s.Name] = s
	}

	// Single-method interface.
	format, ok := byName["Format"]
	if !ok {
		t.Fatal("expected symbol Format (interface method)")
	}
	if format.Kind != symbol.Method {
		t.Errorf("Format.Kind = %v, want Method", format.Kind)
	}
	if format.Parent != "Formatter" {
		t.Errorf("Format.Parent = %q, want Formatter", format.Parent)
	}

	// Multi-method interface with embedded interfaces.
	read, ok := byName["Read"]
	if !ok {
		t.Fatal("expected symbol Read (interface method)")
	}
	if read.Kind != symbol.Method {
		t.Errorf("Read.Kind = %v, want Method", read.Kind)
	}
	if read.Parent != "ReadWriter" {
		t.Errorf("Read.Parent = %q, want ReadWriter", read.Parent)
	}

	write, ok := byName["Write"]
	if !ok {
		t.Fatal("expected symbol Write (interface method)")
	}
	if write.Kind != symbol.Method {
		t.Errorf("Write.Kind = %v, want Method", write.Kind)
	}
	if write.Parent != "ReadWriter" {
		t.Errorf("Write.Parent = %q, want ReadWriter", write.Parent)
	}

	// Embedded interfaces in ReadWriter.
	reader, ok := byName["Reader"]
	if !ok {
		t.Fatal("expected symbol Reader (embedded interface)")
	}
	if reader.Kind != symbol.Type {
		t.Errorf("Reader.Kind = %v, want Type", reader.Kind)
	}
	if reader.Parent != "ReadWriter" {
		t.Errorf("Reader.Parent = %q, want ReadWriter", reader.Parent)
	}

	writer, ok := byName["Writer"]
	if !ok {
		t.Fatal("expected symbol Writer (embedded interface)")
	}
	if writer.Kind != symbol.Type {
		t.Errorf("Writer.Kind = %v, want Type", writer.Kind)
	}
	if writer.Parent != "ReadWriter" {
		t.Errorf("Writer.Parent = %q, want ReadWriter", writer.Parent)
	}
}

func TestParseGoEmbeddedFields(t *testing.T) {
	path := filepath.Join("testdata", "src", "hello.go")
	syms, err := ParseGoFile(path)
	if err != nil {
		t.Fatal(err)
	}

	// Find embedded field symbols — unnamed fields in structs.
	var embedded []symbol.Symbol
	for _, s := range syms {
		if s.Parent == "Node" || s.Parent == "Config" {
			embedded = append(embedded, s)
		}
	}

	// Node embeds *List[int].
	var found bool
	for _, s := range embedded {
		if s.Parent == "Node" && s.Kind == symbol.Type && s.Name == "*List[int]" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected embedded field *List[int] in Node")
	}
}

func TestParseGoFileWithComments(t *testing.T) {
	path := filepath.Join("testdata", "src", "comments.go")
	syms, err := ParseGoFile(path)
	if err != nil {
		t.Fatal(err)
	}

	byName := make(map[string]symbol.Symbol)
	for _, s := range syms {
		byName[s.Name] = s
	}

	doc, ok := byName["DocFunc"]
	if !ok {
		t.Fatal("expected symbol DocFunc")
	}
	if doc.Comment == "" {
		t.Error("DocFunc should have a doc comment")
	}
}

func TestParseEmptyGoFile(t *testing.T) {
	syms, err := ParseGoFile(filepath.Join("testdata", "src", "empty.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(syms) != 0 {
		t.Errorf("got %d symbols from empty file, want 0", len(syms))
	}
}

func TestParseNonExistentGoFile(t *testing.T) {
	_, err := ParseGoFile("/nonexistent/file.go")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestParsePythonFile(t *testing.T) {
	source := `
import os
import sys

# Configuration class
class Config:
    """App configuration."""

    def __init__(self, name):
        self.name = name

    def get_name(self):
        return self.name

# Top-level function
def main():
    config = Config("test")
    return config.get_name()

APP_NAME = "MyApp"
`
	syms, err := ParsePythonFile("test.py", source)
	if err != nil {
		t.Fatal(err)
	}

	byName := make(map[string]symbol.Symbol)
	for _, s := range syms {
		byName[s.Name] = s
	}

	config, ok := byName["Config"]
	if !ok {
		t.Fatal("expected symbol Config")
	}
	if config.Kind != symbol.Class {
		t.Errorf("Config.Kind = %v, want Class", config.Kind)
	}

	init, ok := byName["__init__"]
	if !ok {
		t.Fatal("expected symbol __init__")
	}
	if init.Kind != symbol.Method {
		t.Errorf("__init__.Kind = %v, want Method", init.Kind)
	}
	if init.Parent != "Config" {
		t.Errorf("__init__.Parent = %q, want Config", init.Parent)
	}

	getName, ok := byName["get_name"]
	if !ok {
		t.Fatal("expected symbol get_name")
	}
	if getName.Kind != symbol.Method {
		t.Errorf("get_name.Kind = %v, want Method", getName.Kind)
	}
	if getName.Parent != "Config" {
		t.Errorf("get_name.Parent = %q, want Config", getName.Parent)
	}

	main, ok := byName["main"]
	if !ok {
		t.Fatal("expected symbol main")
	}
	if main.Kind != symbol.Function {
		t.Errorf("main.Kind = %v, want Function", main.Kind)
	}

	appName, ok := byName["AppName"]
	if !ok {
		appName, ok = byName["APP_NAME"]
	}
	if ok {
		if appName.Kind != symbol.Variable {
			t.Errorf("APP_NAME.Kind = %v, want Variable", appName.Kind)
		}
	}
}

func TestParsePythonEmptyFile(t *testing.T) {
	syms, err := ParsePythonFile("empty.py", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(syms) != 0 {
		t.Errorf("got %d symbols from empty file, want 0", len(syms))
	}
}

func TestParsePythonLineNumbers(t *testing.T) {
	source := `def foo():
    pass

def bar():
    x = 1
    y = 2
    return x + y
`
	syms, err := ParsePythonFile("test.py", source)
	if err != nil {
		t.Fatal(err)
	}

	byName := make(map[string]symbol.Symbol)
	for _, s := range syms {
		byName[s.Name] = s
	}

	foo := byName["foo"]
	if foo.LineStart != 1 {
		t.Errorf("foo.LineStart = %d, want 1", foo.LineStart)
	}
	if foo.LineEnd < 2 {
		t.Errorf("foo.LineEnd = %d, want >= 2", foo.LineEnd)
	}

	bar := byName["bar"]
	if bar.LineStart != 4 {
		t.Errorf("bar.LineStart = %d, want 4", bar.LineStart)
	}
	if bar.LineEnd != 7 {
		t.Errorf("bar.LineEnd = %d, want 7", bar.LineEnd)
	}
}

func TestParseGoVariable(t *testing.T) {
	path := filepath.Join("testdata", "src", "hello.go")
	syms, err := ParseGoFile(path)
	if err != nil {
		t.Fatal(err)
	}

	byName := make(map[string]symbol.Symbol)
	for _, s := range syms {
		byName[s.Name] = s
	}

	v, ok := byName["DebugVar"]
	if !ok {
		t.Fatal("expected symbol DebugVar")
	}
	if v.Kind != symbol.Variable {
		t.Errorf("DebugVar.Kind = %v, want Variable", v.Kind)
	}
}
