package helper

type Response struct {
	//StatusCode int         `json:"status_code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func Responses(message string, data interface{}, err interface{}) Response {
	return Response{
		//StatusCode: statusCode,
		Message: message,
		Data:    data,
		Error:   err,
	}
}
