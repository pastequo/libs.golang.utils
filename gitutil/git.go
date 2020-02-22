// Package gitutil allows to get git information at runtime.
package gitutil

// CommitID contains the SHA1 Git commit of the build.
// It's evaluated during compilation.
var CommitID string
