package eval

import (
	"reflect"

	"github.com/Sirupsen/logrus"
)

func complete(v interface{}) error {
	return completeAny(reflect.ValueOf(v))
}

func completeString(v reflect.Value) error {
	newString, err := Value(v.String())
	if err != nil {
		return err
	}

	if newString == v.String() {
		return nil
	}

	logrus.WithFields(logrus.Fields{
		"from": v.String(),
		"to":   newString,
	}).Debug("Explain")

	v.SetString(newString)
	return nil
}

func completeSliceValue(v reflect.Value) error {
	for index := 0; index < v.Len(); index++ {
		if err := completeAny(v.Index(index)); err != nil {
			return err
		}
	}

	return nil
}

func completeStruct(v reflect.Value) error {
	for index := 0; index < v.NumField(); index++ {
		field := v.Field(index)
		if err := completeAny(field); err != nil {
			return err
		}
	}
	return nil
}

func completeAny(v reflect.Value) error {
	if !v.IsValid() {
		return nil
	}

	switch v.Kind() {
	case reflect.String:
		return completeString(v)
	case reflect.Slice:
		return completeSliceValue(v)
	case reflect.Struct:
		return completeStruct(v)
	case reflect.Ptr:
		return completeAny(v.Elem())
	}

	return nil
}
