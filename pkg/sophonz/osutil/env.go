// Package osutil holds small OS/environment helpers shared by sophonz
// components.
package osutil

import "os"

// GetOrDefaultEnv returns the value of the environment variable named key, or
// fallback when the variable is unset or empty.
func GetOrDefaultEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
