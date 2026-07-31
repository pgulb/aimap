package main

import "fmt"

// Greeting returns a friendly message.
func Greeting(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

// Config holds application configuration.
type Config struct {
	Name    string
	Version int
}

// Formatter defines output formatting.
type Formatter interface {
	Format(string) string
}

// GetName returns the config name.
func (c *Config) GetName() string {
	return c.Name
}

// AppVersion is the current version.
const AppVersion = "1.0.0"

// DebugVar controls debug output.
var DebugVar = false

// List is a generic collection.
type List[T any] struct {
	items []T
}

// Push adds an item to the list.
func (l *List[T]) Push(item T) {
	l.items = append(l.items, item)
}

// Pair is a generic pair.
type Pair[T, U any] struct {
	First  T
	Second U
}

// ReadWriter combines io.Reader and io.Writer.
type ReadWriter interface {
	Reader
	Writer
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
}

// NewList creates a new List.
func NewList[T any]() *List[T] {
	return &List[T]{}
}

// Node is a tree node.
type Node struct {
	*List[int]
	Data string
}
