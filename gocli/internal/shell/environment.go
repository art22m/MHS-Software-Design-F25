package shell

import "os"

// NewEnv creates a new Env instance backed by an in-memory map
// for storing and retrieving environment variables.
// It initializes the environment with system environment variables.
func NewEnv() Env {
	env := &envMap{
		store: make(map[string]string),
	}
	for _, pair := range os.Environ() {
		parts := splitEnvPair(pair)
		if len(parts) == 2 {
			env.store[parts[0]] = parts[1]
		}
	}
	return env
}

func splitEnvPair(pair string) []string {
	for i := 0; i < len(pair); i++ {
		if pair[i] == '=' {
			return []string{pair[:i], pair[i+1:]}
		}
	}
	return []string{pair}
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

// GetAll implements Env interface.
// Returns all environment variables as a map.
func (e *envMap) GetAll() map[string]string {
	result := make(map[string]string, len(e.store))
	for k, v := range e.store {
		result[k] = v
	}
	return result
}
