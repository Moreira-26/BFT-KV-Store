package protocol

func CmdNotFoundError(cmd string) string {
	return "Command " + cmd + " not found"
}

func BadArgumentsError() string {
	return "The arguments passed are wrong or badly formatted"
}

func NotYetImplementedError() string {
	return "Not yet implemented"
}
