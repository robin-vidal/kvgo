package resp

// Command represents a pased RESP2 command with its arguments.
type Command struct {
	Name string
	Args []string
}
