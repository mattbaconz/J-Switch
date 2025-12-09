package switcher

// Switch sets the system's Java version to the specified path.
func Switch(javaPath string) error {
	return switchJava(javaPath)
}
