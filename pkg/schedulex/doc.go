// Package schedulex provides a deterministic L1 scheduler API.
//
// Core scheduling decisions are driven by an injectable Clock and Trigger
// values so tests and release evidence can replay behavior without wall-clock
// sleeps. Production code in this package uses only the Go standard library.
package schedulex
