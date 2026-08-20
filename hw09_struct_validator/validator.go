package hw09structvalidator

import (
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

const validateTag = "validate"

var (
	// ErrLen возвращается, если длина строки не совпадает с требуемой.
	ErrLen = fmt.Errorf("invalid length")
	// ErrRegexp возвращается, если строка не соответствует регулярному выражению.
	ErrRegexp = fmt.Errorf("regexp not match")
	// ErrIn возвращается, если значение не входит в допустимое множество.
	ErrIn = fmt.Errorf("value not in set")
	// ErrMin возвращается, если число меньше требуемого минимума.
	ErrMin = fmt.Errorf("value is less than min")
	// ErrMax возвращается, если число больше требуемого максимума.
	ErrMax = fmt.Errorf("value is greater than max")
)

// rulesByKind — поддерживаемые имена правил валидации для каждого типа.
var rulesByKind = map[reflect.Kind][]string{
	reflect.Int:    {"min", "max", "in"},
	reflect.String: {"len", "regexp", "in"},
}

// validateRule — одно правило из тэга validate.
type validateRule struct {
	name  string
	param string
}

// validateFunc валидирует одно значение поля.
type validateFunc func(field reflect.StructField, val reflect.Value) ([]ValidationError, error)

// validatorsByKind — валидаторы значений для каждого типа.
var validatorsByKind = map[reflect.Kind]validateFunc{
	reflect.Int:    validateInt,
	reflect.String: validateString,
}

// ValidationError — ошибка валидации одного поля.
type ValidationError struct {
	Field string
	Err   error
}

// Error возвращает имя поля и ошибку его валидации.
func (v ValidationError) Error() string {
	return fmt.Sprintf("%s: %v", v.Field, v.Err)
}

// ValidationErrors — список ошибок валидации полей.
type ValidationErrors []ValidationError

// Error возвращает все ошибки валидации полей одной строкой.
func (v ValidationErrors) Error() string {
	msgs := make([]string, 0, len(v))
	for _, fieldErr := range v {
		msgs = append(msgs, fieldErr.Error())
	}

	return strings.Join(msgs, "; ")
}

// Validate валидирует публичные поля структуры по тэгам validate.
func Validate(v interface{}) error {
	if v == nil {
		return fmt.Errorf("v is nil")
	}

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("v must be a struct")
	}

	typ := val.Type()
	errs := make(ValidationErrors, 0)

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldErrs, err := validateField(field, val.Field(i))
		if err != nil {
			return fmt.Errorf("validate field %s: %w", field.Name, err)
		}

		errs = append(errs, fieldErrs...)
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

// validateField валидирует одно поле структуры.
func validateField(field reflect.StructField, val reflect.Value) ([]ValidationError, error) {
	_, ok := field.Tag.Lookup(validateTag)
	if !ok {
		return nil, nil
	}

	if field.Type.Kind() == reflect.Slice {
		return validateSlice(field, val)
	}

	return validateValue(field, val)
}

// validateValue выбирает валидатор по типу значения.
func validateValue(field reflect.StructField, val reflect.Value) ([]ValidationError, error) {
	fn, ok := validatorsByKind[val.Kind()]
	if !ok {
		return nil, fmt.Errorf("unsupported type: %s", val.Kind())
	}

	return fn(field, val)
}

// validateSlice валидирует каждый элемент слайса.
func validateSlice(field reflect.StructField, val reflect.Value) ([]ValidationError, error) {
	elemKind := val.Type().Elem().Kind()
	if _, ok := validatorsByKind[elemKind]; !ok {
		return nil, fmt.Errorf("unsupported slice type: %s", elemKind)
	}

	errs := make(ValidationErrors, 0)

	for i := 0; i < val.Len(); i++ {
		elemErrs, err := validateValue(field, val.Index(i))
		if err != nil {
			return nil, err
		}

		errs = append(errs, elemErrs...)
	}

	return errs, nil
}

// validateInt валидирует целочисленное значение по правилам тэга.
func validateInt(field reflect.StructField, val reflect.Value) ([]ValidationError, error) {
	allowed, ok := rulesByKind[reflect.Int]
	if !ok {
		return nil, fmt.Errorf("unsupported type: %s", val.Kind())
	}

	rules, err := parseRules(field.Tag.Get(validateTag))
	if err != nil {
		return nil, err
	}

	errs := make(ValidationErrors, 0)
	n := val.Int()

	for _, rule := range rules {
		if !slices.Contains(allowed, rule.name) {
			return nil, fmt.Errorf("unexpected rule %q for int", rule.name)
		}

		fieldErr, err := applyIntRule(field.Name, n, rule)
		if err != nil {
			return nil, err
		}

		if fieldErr != nil {
			errs = append(errs, *fieldErr)
		}
	}

	return errs, nil
}

// validateString валидирует строковое значение по правилам тэга.
func validateString(field reflect.StructField, val reflect.Value) ([]ValidationError, error) {
	allowed, ok := rulesByKind[reflect.String]
	if !ok {
		return nil, fmt.Errorf("unsupported type: %s", val.Kind())
	}

	rules, err := parseRules(field.Tag.Get(validateTag))
	if err != nil {
		return nil, err
	}

	errs := make(ValidationErrors, 0)
	s := val.String()

	for _, rule := range rules {
		if !slices.Contains(allowed, rule.name) {
			return nil, fmt.Errorf("unexpected rule %q for string", rule.name)
		}

		fieldErr, err := applyStringRule(field.Name, s, rule)
		if err != nil {
			return nil, err
		}

		if fieldErr != nil {
			errs = append(errs, *fieldErr)
		}
	}

	return errs, nil
}

// parseRules разбирает тэг validate на отдельные правила.
func parseRules(tagValue string) ([]validateRule, error) {
	parts := strings.Split(tagValue, "|")
	rules := make([]validateRule, 0, len(parts))

	for _, part := range parts {
		name, param, ok := strings.Cut(part, ":")
		if !ok || name == "" || param == "" {
			return nil, fmt.Errorf("invalid validate tag %q", tagValue)
		}

		rules = append(rules, validateRule{name: name, param: param})
	}

	return rules, nil
}

// applyIntRule применяет одно правило валидации к целому числу.
func applyIntRule(fieldName string, n int64, rule validateRule) (*ValidationError, error) {
	switch rule.name {
	case "min":
		minVal, err := strconv.ParseInt(rule.param, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid min value %q: %w", rule.param, err)
		}

		if n < minVal {
			return &ValidationError{Field: fieldName, Err: fmt.Errorf("%w", ErrMin)}, nil
		}
	case "max":
		maxVal, err := strconv.ParseInt(rule.param, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid max value %q: %w", rule.param, err)
		}

		if n > maxVal {
			return &ValidationError{Field: fieldName, Err: fmt.Errorf("%w", ErrMax)}, nil
		}
	case "in":
		ok, err := intInSet(n, rule.param)
		if err != nil {
			return nil, err
		}

		if !ok {
			return &ValidationError{Field: fieldName, Err: fmt.Errorf("%w", ErrIn)}, nil
		}
	}

	return nil, nil
}

// applyStringRule применяет одно правило валидации к строке.
func applyStringRule(fieldName string, s string, rule validateRule) (*ValidationError, error) {
	switch rule.name {
	case "len":
		wantLen, err := strconv.Atoi(rule.param)
		if err != nil {
			return nil, fmt.Errorf("invalid len value %q: %w", rule.param, err)
		}

		if utf8.RuneCountInString(s) != wantLen {
			return &ValidationError{Field: fieldName, Err: fmt.Errorf("%w", ErrLen)}, nil
		}
	case "regexp":
		re, err := regexp.Compile(rule.param)
		if err != nil {
			return nil, fmt.Errorf("invalid regexp %q: %w", rule.param, err)
		}

		if !re.MatchString(s) {
			return &ValidationError{Field: fieldName, Err: fmt.Errorf("%w", ErrRegexp)}, nil
		}
	case "in":
		set := strings.Split(rule.param, ",")
		if !slices.Contains(set, s) {
			return &ValidationError{Field: fieldName, Err: fmt.Errorf("%w", ErrIn)}, nil
		}
	}

	return nil, nil
}

// intInSet проверяет, входит ли число в множество правила in.
func intInSet(n int64, param string) (bool, error) {
	parts := strings.Split(param, ",")

	for _, part := range parts {
		item, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return false, fmt.Errorf("invalid in value %q: %w", part, err)
		}

		if item == n {
			return true, nil
		}
	}

	return false, nil
}
