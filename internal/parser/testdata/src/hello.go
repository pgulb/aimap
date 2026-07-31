// Package main provides a sample program for parser tests.
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
