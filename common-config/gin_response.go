package common

import (
	"errors"
	"net/http"

	common_errors "github.com/weiqiangxu/common-config/error_code"

	"github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"
)

var DefaultValidator = validator.New()

// SuccessDto 成功数据 结构体
type SuccessDto struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

// FailDto 失败数据 结构体
type FailDto struct {
	Code  int      `json:"code"` // 状态码
	Error ErrorDto `json:"error"`
}

// ErrorDto 失败数据-错误 结构体
type ErrorDto struct {
	Code    string `json:"code"`    // 错误码
	Message string `json:"message"` // 错误描述
}

// Page 统一页码结构
type Page struct {
	Count     int `json:"count"`
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
}

// PageResponse 分页响应
type PageResponse struct {
	Page
	List interface{} `json:"list"`
}

// ResponseSuccess 返回成功
func ResponseSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, SuccessDto{Code: common_errors.CodeSuccess, Data: data})
}

// ResponseInvalidParams 返回无效参数
func ResponseInvalidParams(c *gin.Context, err error) {
	c.JSON(http.StatusOK, FailDto{
		Code: common_errors.CodeInvalidParams,
		Error: ErrorDto{
			Code:    common_errors.CodeInvalidParamsMessage,
			Message: err.Error(),
		},
	})
}

func ResponsePage(c *gin.Context, list interface{}, count int, pageIndex int, pageSize int) {
	res := PageResponse{
		Page: Page{
			Count:     count,
			PageIndex: pageIndex,
			PageSize:  pageSize,
		},
		List: list,
	}
	c.JSON(http.StatusOK, SuccessDto{
		Code: common_errors.CodeSuccess,
		Data: res,
	})
}

// ResponseError 返回错误
func ResponseError(c *gin.Context, code int, codeStr string, err error) {
	c.JSON(http.StatusOK, FailDto{Code: code, Error: ErrorDto{Code: codeStr, Message: err.Error()}})
}

// ResponseEncryptSuccess 加密返回
func ResponseEncryptSuccess(c *gin.Context, data interface{}) {
}

// ResponseEncryptPage 加密分页
func ResponseEncryptPage(c *gin.Context, list interface{}, count int, pageIndex int, pageSize int) {
}

func BindAndValid(c *gin.Context, form interface{}) error {
	err := c.Bind(form)
	if err != nil {
		return err
	}
	err = DefaultValidator.Struct(form)

	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return validationErrors
		}
		return err
	}
	return nil
}
