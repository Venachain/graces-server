package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	TimeTemplate = "2006-01-02 15:04:05"
)

// SimpleCopyProperties 反射浅拷贝
// 注意：此函数的拷贝在遇到引用类型时，返回的是原引用数据，
// 也就是说，拷贝后的 dst 在对引用类型的数据进行修改时，会影响到 src 对应的数据
func SimpleCopyProperties(dst, src interface{}) (err error) {
	// 防止意外panic
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprintf("%v", e))
		}
	}()

	if dst == nil {
		return errors.New("dst not be nil")
	}

	dstType, dstValue := reflect.TypeOf(dst), reflect.ValueOf(dst)
	srcType, srcValue := reflect.TypeOf(src), reflect.ValueOf(src)

	// dst必须结构体指针类型
	if dstType.Kind() != reflect.Ptr || dstType.Elem().Kind() != reflect.Struct {
		return errors.New("dst type should be a struct pointer")
	}

	// src必须为结构体或者结构体指针，.Elem()类似于*ptr的操作返回指针指向的地址反射类型
	if srcType.Kind() == reflect.Ptr {
		srcType, srcValue = srcType.Elem(), srcValue.Elem()
	}
	if srcType.Kind() != reflect.Struct {
		return errors.New("src type should be a struct or a struct pointer")
	}

	// 取具体内容
	dstType, dstValue = dstType.Elem(), dstValue.Elem()

	// 属性个数
	propertyNums := dstType.NumField()
	for i := 0; i < propertyNums; i++ {
		// 属性
		property := dstType.Field(i)
		// 待填充属性值property = {reflect.StructField}
		propertyValue := srcValue.FieldByName(property.Name)

		// 无效，说明src没有这个属性 || 属性同名但类型不同
		if !propertyValue.IsValid() || property.Type != propertyValue.Type() {
			continue
		}
		// 只能对可导出字段进行赋值
		if dstValue.Field(i).CanSet() {
			dstValue.Field(i).Set(propertyValue)
		}
	}
	return nil
}

//Clone 浅克隆，可以克隆任意数据类型，对指针类型子元素无法克隆
//获取类型：如果类型是指针类型，需要使用Elem()获取对象实际类型
//获取实际值：如果值是指针类型，需要使用Elem()获取实际数据
func Clone(src interface{}) interface{} {
	typ := reflect.TypeOf(src)
	if typ.Kind() == reflect.Ptr { //如果是指针类型
		typ = typ.Elem()               //获取源实际类型(否则为指针类型)
		dst := reflect.New(typ).Elem() //创建对象
		data := reflect.ValueOf(src)   //源数据值
		data = data.Elem()             //源数据实际值（否则为指针）
		dst.Set(data)                  //设置数据
		dst = dst.Addr()               //创建对象的地址（否则返回值）
		return dst.Interface()         //返回地址
	} else {
		dst := reflect.New(typ).Elem() //创建对象
		data := reflect.ValueOf(src)   //源数据值
		dst.Set(data)                  //设置数据
		return dst.Interface()         //返回
	}
}

//DeepClone 深度克隆，可以克隆任意数据类型
func DeepClone(src interface{}) interface{} {
	typ := reflect.TypeOf(src)
	if typ.Kind() == reflect.Ptr { //如果是指针类型
		typ = typ.Elem()                              //获取源实际类型(否则为指针类型)
		dst := reflect.New(typ).Elem()                //创建对象
		b, _ := json.Marshal(src)                     //导出json
		_ = json.Unmarshal(b, dst.Addr().Interface()) //json序列化
		return dst.Addr().Interface()                 //返回指针
	} else {
		dst := reflect.New(typ).Elem()                //创建对象
		b, _ := json.Marshal(src)                     //导出json
		_ = json.Unmarshal(b, dst.Addr().Interface()) //json序列化
		return dst.Interface()                        //返回值
	}
}

func BuildOptionsByQuery(pageIndex, pageSize int64) *options.FindOptions {
	if pageIndex <= 0 {
		pageIndex = 1
	}
	findOps := options.Find()
	findOps.SetSkip((pageIndex - 1) * pageSize)
	findOps.SetLimit(pageSize)

	return findOps
}

func Timestamp2TimeStr(timestamp int64) string {
	if timestamp <= 0 {
		return "-"
	}
	return time.Unix(timestamp, 0).Format(TimeTemplate)
}

// Hex2Number 16 进制转 10 进制
func Hex2Number(hex string) (uint64, error) {
	val := hex[2:]
	n, err := strconv.ParseUint(val, 16, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// MethodCapitalized 方法名首字母大写
func MethodCapitalized(methodName string) string {
	methodNameBytes := []byte(methodName)
	if methodNameBytes[0] >= 'a' && methodNameBytes[0] <= 'z' {
		// 首字母大写
		methodNameBytes[0] = methodNameBytes[0] - 32
	}
	return string(methodNameBytes)
}

func JsonStringPatch(value interface{}) (interface{}, error) {
	if reflect.ValueOf(value).Kind() == reflect.String {
		str := value.(string)
		// string starts with {"
		res := bytes.Trim([]byte(str), "\u0000")
		if bytes.Equal([]byte(str)[:2], []byte{123, 34}) {
			var m map[string]interface{}
			b := bytes.Trim([]byte(str), "\x00")
			//c := bytes.Trim([]byte(b), "\u0000")
			err := json.Unmarshal(b, &m)
			if err != nil {
				return nil, err
			}
			return m, nil
		}
		return string(res), nil
	}
	return value, nil
}
