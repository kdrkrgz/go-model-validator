package main

import (
	"errors"
	"log"
	"reflect"
	"strings"
	"unicode"
)

func main() {
	product := Product{
		ProductID:   1,
		ProductName: "TestProduct",
		Quantity:    1,
		IsActive:    true,
		Slug:        "test",
	}
	//var a interface{}
	//b := 1
	validator := IFieldValidator(&FieldValidator{product, true})
	var _ = validator.Validate()
}

type Product struct {
	ProductID   int    `json:"productId" validate:"required,positiveNumberField"`
	ProductName string `json:"productName" validate:"nameField"`
	Quantity    int    `json:"quantity" validate:"positiveNumberField"`
	IsActive    bool   `json:"isActive" validate:"alwaysTrueField"`
	Slug        string `json:"slug" validate:"slugField"`
}

// Get field and tag names from model
func GetFieldAndTagNames(obj any) map[string][]string {
	useStructFieldNamesForValidation := true
	fieldsAndTags := make(map[string][]string)
	t := reflect.Indirect(reflect.ValueOf(obj)).Type()
	for i := 0; i < t.NumField(); i++ {
		var field string
		field = t.Field(i).Tag.Get("json")
		if useStructFieldNamesForValidation {
			field = t.Field(i).Name
		}
		tag := t.Field(i).Tag.Get("validate")
		fieldsAndTags[field] = append(fieldsAndTags[field], tag)
	}
	return fieldsAndTags
}

type IFieldValidator interface {
	Validate() error
}

type FieldValidator struct {
	model        any
	failSilently bool
}

func (validator *FieldValidator) Validate() error {
	// get tags and field from model map[string][]string
	rv := reflect.ValueOf(validator.model)
	if rv.Kind() != reflect.Struct {
		log.Panicf("Non struct types can not validate: %v", rv.Kind())
	}
	tags := GetFieldAndTagNames(validator.model)
	for k, v := range tags {
		// get model field values
		fieldValue := rv.FieldByName(k)
		for _, tag := range v {
			// split tags
			splittedTags := strings.Split(tag, ",")
			for _, splittedTag := range splittedTags {
				// get validator map
				validatorMap := getValidatorMap()
				// get validator function by tag
				validatorFunc, ok := validatorMap[splittedTag]
				if !ok {
					log.Fatalf("Unknown tag: %s", splittedTag)
				}
				// call validator function with field value
				validatorFunc(fieldValue.Interface(), validator.failSilently)
			}

		}
	}
	return nil
}

// validation functions
func positiveNumberFieldValidator(v any, failSilently bool) {
	if v.(int) < 0 {
		err := newValidationError(errors.New("positive number field must greater then zero"), failSilently)
		throwError(err)
	}
}

func alwaysTrueValidator(v any, failSilently bool) {
	if v.(bool) != true {
		err := newValidationError(errors.New("always true field must be true"), failSilently)
		throwError(err)
	}
}

func nameFieldValidator(v any, failSilently bool) {
	if len(strings.Split(v.(string), " ")) == 1 {
		err := newValidationError(errors.New("name field must be minimum 2 words"), failSilently)
		throwError(err)
	}
}

func slugFieldValidator(v any, failSilently bool) {
	hasUpper := false
	for _, c := range v.(string) {
		if unicode.IsUpper(c) {
			hasUpper = true
			break
		}
	}
	if hasUpper {
		err := newValidationError(errors.New("slug field must contains lowercase letters"), failSilently)
		throwError(err)
	}
}

func requiredFieldValidator(v any, failSilently bool) {
	if v == nil {
		err := newValidationError(errors.New("required field cannot be empty"), failSilently)
		throwError(err)
	}
}

type ValidationError struct {
	err      error
	isSilent bool
}

func newValidationError(err error, isSilent bool) ValidationError {
	return ValidationError{
		err:      err,
		isSilent: isSilent,
	}
}

func throwError(err ValidationError) {
	if !err.isSilent {
		log.Panicln(err)
	}
	log.Println(err.err)
}

func getValidatorMap() map[string]func(v any, failSilently bool) {
	validatorMap := map[string]func(v any, failSilently bool){
		"required":            requiredFieldValidator,
		"positiveNumberField": positiveNumberFieldValidator,
		"nameField":           nameFieldValidator,
		"alwaysTrueField":     alwaysTrueValidator,
		"slugField":           slugFieldValidator,
	}
	return validatorMap
}
