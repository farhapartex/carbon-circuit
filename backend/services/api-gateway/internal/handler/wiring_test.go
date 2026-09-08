package handler

import (
	"reflect"
	"testing"
)

func filledOptions(t *testing.T) RouterOptions {
	t.Helper()

	options := RouterOptions{}
	value := reflect.ValueOf(&options).Elem()

	for index := range value.NumField() {
		field := value.Field(index)
		if field.Kind() == reflect.Pointer && field.CanSet() {
			field.Set(reflect.New(field.Type().Elem()))
		}
	}

	return options
}

func TestEveryUpstreamOfferedByRouterOptionsReachesTheHandlers(t *testing.T) {
	options := filledOptions(t)
	handlers := reflect.ValueOf(handlersFrom(options)).Elem()

	offered := reflect.TypeOf(RouterOptions{})
	checked := 0

	for index := range handlers.NumField() {
		field := handlers.Type().Field(index)

		if field.Type.Kind() != reflect.Pointer {
			continue
		}
		if _, present := offered.FieldByName(field.Name); !present {
			continue
		}

		checked++

		if handlers.Field(index).IsNil() {
			t.Errorf(
				"RouterOptions offers %s but handlersFrom never assigns it, so the first request through a handler that uses it panics",
				field.Name,
			)
		}
	}

	if checked < 5 {
		t.Fatalf("expected to check several upstreams, only found %d", checked)
	}
}

func TestSessionDenylistIsCarriedUnderItsOtherName(t *testing.T) {
	options := filledOptions(t)
	handlers := handlersFrom(options)

	if handlers.Denylist == nil {
		t.Fatal("Handlers.Denylist is fed by RouterOptions.SessionDenylist, and was not assigned")
	}
}
