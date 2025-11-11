package shell

// NewEnv creates a new Env instance backed by an in-memory map
// for storing and retrieving environment variables.
func NewEnv() Env {
	return &envMap{
		store: make(map[string]string),
	}
}

type envMap struct {
	store map[string]string
}

// Get implements Env interface.
// Retrieves the value associated with the given key from the environment store.
func (e *envMap) Get(key string) (value string, ok bool) {
	value, ok = e.store[key]
	return
}

// Set implements Env interface.
// Stores a key-value pair in the environment.
func (e *envMap) Set(key string, value string) {
	e.store[key] = value
}
