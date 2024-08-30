package api

type ResultCode int

const (
	Success ResultCode = 0
	Error   ResultCode = -1
)

type JSONResult struct {
	Code ResultCode  `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func RSuccess(data interface{}) JSONResult {
	return JSONResult{
		Code: Success,
		Msg:  "success",
		Data: data,
	}
}

func RError(msg string) JSONResult {
	return JSONResult{
		Code: Error,
		Msg:  msg,
		Data: nil,
	}
}
