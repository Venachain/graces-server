package validate

import (
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
	zhTrans "gopkg.in/go-playground/validator.v9/translations/zh"
)

var (
	// myValidator 数据校验器
	myValidator *validator.Validate
	// zHCNTranslator 中文翻译器
	zHCNTranslator ut.Translator
)

func init() {
	myValidator = validator.New()
	zhCn := zh.New()
	uni := ut.New(zhCn)

	zHCNTranslator, _ = uni.GetTranslator("zh")
	err := zhTrans.RegisterDefaultTranslations(myValidator, zHCNTranslator)
	if err != nil {
		log.Fatalln(err)
	}
}

// Validate 数据校验
// 返回校验失败时已格式化了的错误信息
func Validate(obj interface{}) error {
	reType := reflect.TypeOf(obj)
	if reType.Kind() == reflect.Ptr {
		reType = reType.Elem()
	}
	if reType.Kind() != reflect.Struct {
		str := "被校验的对象必须是一个结构体"
		return errors.New(str)
	}
	err := myValidator.Struct(obj)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); !ok {
			if errs, ok := err.(validator.ValidationErrors); ok {
				for _, e := range errs {
					str := fmt.Sprintf("%s:%s \n", e.StructNamespace(), e.Translate(zHCNTranslator))
					return errors.New(str)
				}
			}
		}
	}
	return nil
}
