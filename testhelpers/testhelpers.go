package testhelpers

import (
	"net/http/httptest"
	"testing"
)

func CheckSlicesEquality(a []any, b []any) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[any]int)
	bMap := make(map[any]int)

	for _, v := range a {
		aMap[v]++
	}
	for _, v := range b {
		bMap[v]++
	}

	for k, v := range aMap {
		if bMap[k] != v {
			return false
		}
	}

	return true
}

func StringSliceToAnySlice(s []string) []any {
	var result []any
	for _, v := range s {
		result = append(result, v)
	}
	return result
}

type ExpectedResponse struct {
	Code   int
	Header string
	Body   string
}

func VerifyResponse(t *testing.T, rr *httptest.ResponseRecorder, expectedResponse ExpectedResponse) {
	t.Helper()

	expectedCode := expectedResponse.Code
	expectedHeader := expectedResponse.Header
	expectedBody := expectedResponse.Body

	if status := rr.Code; status != expectedCode {
		t.Errorf("Expected status code: %d, Actual: %d", expectedCode, status)
	}

	if header := rr.Header().Get("Content-Type"); header != expectedHeader {
		t.Errorf("Expected Content-Type: %s, Actual: %s", expectedHeader, header)
	}

	if body := rr.Body.String(); body != expectedBody {
		t.Errorf("Expected body: %q, Actual: %q", expectedBody, body)
	}
}
