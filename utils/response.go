package utils

import "github.com/gin-gonic/gin"

type BaseResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Object  interface{} `json:"object,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

func JSON(c *gin.Context, status int, success bool, msg string, obj interface{}, errs interface{}) {
	c.JSON(status, BaseResponse{Success: success, Message: msg, Object: obj, Errors: errs})
}
