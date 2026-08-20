package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

type UserRole string

// Структуры для проверки валидации разных типов полей.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}

	Numbers struct {
		Values []int `validate:"min:10|max:20"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		in       interface{}
		wantErrs []ValidationError
	}{
		{
			name: "valid user",
			in: User{
				ID:     "123e4567-e89b-12d3-a456-426614174000",
				Name:   "John",
				Age:    25,
				Email:  "user@mail.com",
				Role:   "admin",
				Phones: []string{"79001234567"},
			},
		},
		{
			name: "token without validate tags",
			in: Token{
				Header:    []byte{1},
				Payload:   []byte{2},
				Signature: []byte{3},
			},
		},
		{
			name: "valid app",
			in:   App{Version: "1.1.0"},
		},
		{
			name: "valid response",
			in:   Response{Code: 200, Body: "ok"},
		},
		{
			name: "valid numbers slice",
			in:   Numbers{Values: []int{10, 15, 20}},
		},
		{
			name: "invalid app length",
			in:   App{Version: "1.1"},
			wantErrs: []ValidationError{
				{Field: "Version", Err: ErrLen},
			},
		},
		{
			name: "invalid response code",
			in:   Response{Code: 201, Body: "created"},
			wantErrs: []ValidationError{
				{Field: "Code", Err: ErrIn},
			},
		},
		{
			name: "user age less than min",
			in: User{
				ID:     "123e4567-e89b-12d3-a456-426614174000",
				Name:   "John",
				Age:    10,
				Email:  "user@mail.com",
				Role:   "admin",
				Phones: []string{"79001234567"},
			},
			wantErrs: []ValidationError{
				{Field: "Age", Err: ErrMin},
			},
		},
		{
			name: "user age greater than max",
			in: User{
				ID:     "123e4567-e89b-12d3-a456-426614174000",
				Name:   "John",
				Age:    60,
				Email:  "user@mail.com",
				Role:   "admin",
				Phones: []string{"79001234567"},
			},
			wantErrs: []ValidationError{
				{Field: "Age", Err: ErrMax},
			},
		},
		{
			name: "multiple user errors",
			in: User{
				ID:     "short",
				Name:   "John",
				Age:    10,
				Email:  "bad-email",
				Role:   "user",
				Phones: []string{"79001234567", "123"},
			},
			wantErrs: []ValidationError{
				{Field: "ID", Err: ErrLen},
				{Field: "Age", Err: ErrMin},
				{Field: "Email", Err: ErrRegexp},
				{Field: "Role", Err: ErrIn},
				{Field: "Phones", Err: ErrLen},
			},
		},
		{
			name: "numbers slice validates each element",
			in:   Numbers{Values: []int{5, 15, 25}},
			wantErrs: []ValidationError{
				{Field: "Values", Err: ErrMin},
				{Field: "Values", Err: ErrMax},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d %s", i, tt.name), func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tt.in)
			assertValidationErrors(t, err, tt.wantErrs)
		})
	}
}

func TestValidateProgramErrors(t *testing.T) {
	t.Parallel()

	type invalidRegexp struct {
		Value string `validate:"regexp:["`
	}
	type minOnString struct {
		Value string `validate:"min:10"`
	}
	type invalidTag struct {
		Value string `validate:"len"`
	}
	type unsupported struct {
		Value bool `validate:"len:1"`
	}

	tests := []struct {
		name string
		in   interface{}
	}{
		{name: "nil", in: nil},
		{name: "not a struct", in: 42},
		{name: "invalid regexp", in: invalidRegexp{Value: "123"}},
		{name: "unexpected validator", in: minOnString{Value: "abc"}},
		{name: "invalid tag format", in: invalidTag{Value: "abc"}},
		{name: "unsupported type", in: unsupported{Value: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tt.in)
			if err == nil {
				t.Fatal("expected program error, got nil")
			}

			var vErrs ValidationErrors
			if errors.As(err, &vErrs) {
				t.Fatalf("expected program error, got ValidationErrors: %v", err)
			}
		})
	}
}

func assertValidationErrors(t *testing.T, err error, want []ValidationError) {
	t.Helper()

	if len(want) == 0 {
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		return
	}

	var got ValidationErrors
	if !errors.As(err, &got) {
		t.Fatalf("expected ValidationErrors, got %v (%T)", err, err)
	}

	if len(got) != len(want) {
		t.Fatalf("got %d errors (%v), want %d (%v)", len(got), got, len(want), want)
	}

	for i := range want {
		if got[i].Field != want[i].Field {
			t.Errorf("error[%d].Field = %q, want %q", i, got[i].Field, want[i].Field)
		}

		if !errors.Is(got[i].Err, want[i].Err) {
			t.Errorf("error[%d].Err = %v, want %v", i, got[i].Err, want[i].Err)
		}
	}
}
