package bid754

// NativeBackendEnabled reports whether the CGO-backed decimal implementations
// are compiled into the current build.
func NativeBackendEnabled() bool {
	return nativeBackendEnabled
}
