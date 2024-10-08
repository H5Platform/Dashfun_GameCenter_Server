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

func PageSuccess[T any](data T, page, size int64, totalPages int) JSONResult {
	return RSuccess(&PageResult[T]{
		Data:       data,
		Page:       page,
		Size:       size,
		TotalPages: totalPages,
	})
}

type PageResult[T any] struct {
	TotalPages int   `json:"total_pages"`
	Page       int64 `json:"page"`
	Size       int64 `json:"size"`
	Data       T     `json:"data"`
}
