package ron

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

type FooBody struct {
	FString  string    `json:"fstring" form:"fstring"`
	FInt     int       `json:"fint" form:"fint"`
	FUint    uint      `json:"fuint" form:"fuint"`
	FFloat64 float64   `json:"ffloat64" form:"ffloat64"`
	FBool    bool      `json:"fbool" form:"fbool"`
	FTime    time.Time `json:"ftime" form:"ftime"`

	FStringSlice []string `json:"fstring_slice" form:"fstring_slice"`
	FIntSlice    []int    `json:"fint_slice" form:"fint_slice"`
	FBoolSlice   []bool   `json:"fbool_slice" form:"fbool_slice"`

	FNone string

	fPrivate string
}

func expectedStruct() (FooBody, time.Time) {
	actualTime := time.Now().Round(time.Second).UTC()
	fooBody := FooBody{
		FString:  "string",
		FInt:     -30,
		FUint:    30,
		FFloat64: 3.14,
		FBool:    true,
		FTime:    actualTime,

		FStringSlice: []string{"string1", "string2"},
		FIntSlice:    []int{1, 2},
		FBoolSlice:   []bool{true, false},
	}

	return fooBody, actualTime
}

func Test_BindJSON(t *testing.T) {
	expected, _ := expectedStruct()
	body, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("Marshal() failed: %v", err)
	}

	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c := &CTX{
		W: rr,
		R: req,
	}

	var foo FooBody
	err = c.BindJSON(&foo)
	if err != nil {
		t.Errorf("BindJSON() failed: %v", err)
	}

	if reflect.DeepEqual(foo, expected) == false {
		t.Errorf("Expected: %v, Actual: %v", expected, foo)
	}
}

func Test_BindForm(t *testing.T) {
	expected, actualTime := expectedStruct()
	req := httptest.NewRequest("POST", "/", nil)
	req.Form = map[string][]string{
		"fstring":       {"string"},
		"fint":          {"-30"},
		"fuint":         {"30"},
		"ffloat64":      {"3.14"},
		"fbool":         {"true"},
		"ftime":         {actualTime.Format(time.RFC3339)},
		"fstring_slice": {"string1", "string2"},
		"fint_slice":    {"1", "2"},
		"fbool_slice":   {"true", "false"},
		"fnone":         {"none"},
	}

	rr := httptest.NewRecorder()

	c := &CTX{
		W: rr,
		R: req,
	}

	var foo FooBody
	err := c.BindForm(&foo)
	if err != nil {
		t.Errorf("BindForm() failed: %v", err)
	}

	if reflect.DeepEqual(foo, expected) == false {
		t.Errorf("Expected: %v, Actual: %v", expected, foo)
	}
}
