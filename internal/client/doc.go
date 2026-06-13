// Package client is the Anthropic API boundary: HTTP transport, retry, and
// response parsing (including usage accounting and truncation handling).
//
// It is the ONLY place a test fake for the API lives; no other package stubs
// the network. Callers pass a job and config in and get a parsed result out.
package client
