// Package messages defines various message types used within the TUI application.
// These messages are used to communicate between different components and handle
// various actions such as opening and closing configuration models, setting information
// messages, and handling errors.
package messages

// Error represents an error message.
type Error struct {
	Err error
}

// Info represents an informational message with optional help text.
type Info struct {
	Info string
	Help string
}

// OpenConfigModel represents a message to open the configuration model.
type OpenConfigModel struct {
}

// CloseConfigModel represents a message to close the configuration model.
// It includes the path to the key and the address to the service.
type CloseConfigModel struct {
	PathToKey     string
	AddrToService string
}

// Registration represents a message to handle user registration.
// It includes the necessary details for registering a new user.
type Registration struct {
	Login    string
	Password string
}

// SignIn represents a message to handle user sign-in.
// It includes the necessary details for authenticating a user.
type SignIn struct {
	Login    string
	Password string
}
