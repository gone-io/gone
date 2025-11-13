package gone

// Default is a singleton instance of Application, used to simplify common usage patterns.
var Default = NewApp()

// Loads uses the default application instance to load multiple LoadFunc functions.
// Parameters:
//   - loads: One or more LoadFunc functions that will be executed in sequence
//
// Returns:
//   - *Application: Returns the default application instance for method chaining
func Loads(loads ...LoadFunc) *Application {
	return Default.Loads(loads...)
}

// Load uses the default application instance to load a Goner with optional configuration options.
// Parameters:
//   - goner: The Goner instance to load
//   - options: Optional configuration options for the Goner
//
// Returns:
//   - *Application: Returns the default application instance for method chaining
func Load(goner Goner, options ...Option) *Application {
	return Default.Load(goner, options...)
}

// Run executes one or more functions using the default application instance.
// These functions can have dependencies that will be automatically injected.
// Parameters:
//   - fn: A variadic list of functions to execute with injected dependencies
func Run(fn ...any) {
	Default.Run(fn...)
}

// Serve starts all daemons and waits for termination signal using the default application instance.
// This function will start all registered daemons and block until a shutdown signal is received.
func Serve(fn ...any) {
	Default.Serve(fn...)
}

// End triggers application termination
// It terminates the application by sending a SIGINT signal to the default Application instance
// This is a convenience method equivalent to calling Default.End()
func End() {
	Default.End()
}

// Test runs tests using the default application instance.
// Parameters:
//   - fn: The test function to execute with injected dependencies
func Test(fn any) {
	Default.Test(fn)
}
