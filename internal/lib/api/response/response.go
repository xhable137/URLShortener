package response

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func Error(msg string) Response {
	return Response{
		Status: "Error",
		Error:  msg,
	}
}

func OK() Response {
	return Response{
		Status: "OK",
	}
}
