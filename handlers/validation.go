package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// formatValidationError 把「綁定／驗證請求 body 時產生的錯誤」整理成 API 要回傳的格式。
func formatValidationError(err error) (int, interface{}) {
	if err == nil {
		return 0, nil
	}
	if errs, ok := err.(validator.ValidationErrors); ok {
		details := make([]gin.H, 0, len(errs))
		for _, e := range errs {
			field := e.Field()
			if len(field) > 0 {
				field = strings.ToLower(field[:1]) + field[1:]
			}
			details = append(details, gin.H{
				"field":   field,
				"message": validationMessage(e.Tag(), e.Param()),
			})
		}
		return http.StatusBadRequest, gin.H{
			"error":   "資料驗證失敗",
			"details": details,
		}
	}
	return http.StatusBadRequest, gin.H{"error": err.Error()}
}

func validationMessage(tag, param string) string {
	switch tag {
	case "required":
		return "此欄位為必填"
	case "max":
		return "超過最大長度 " + param
	case "min":
		return "低於最小值 " + param
	case "gte":
		return "須大於等於 " + param
	case "lte":
		return "須小於等於 " + param
	default:
		return "不符合規則: " + tag
	}
}
