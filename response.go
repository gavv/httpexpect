package httpexpect

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"mime"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type Response struct {
	testing *testing.T
	resp    *http.Response
}

func (r *Response) ExpectCode(code int) *Response {
	require.Equal(r.testing, r.resp.StatusCode, code)
	return r
}

func (r *Response) ExpectEmpty() *Response {
	contentType := r.resp.Header.Get("Content-Type")

	require.Equal(r.testing, contentType, "")
	require.Equal(r.testing, r.getString(), "")

	return r
}

func (r *Response) ExpectJSON() *Response {
	contentType := r.resp.Header.Get("Content-Type")

	mediaType, params, _ := mime.ParseMediaType(contentType)
	charset := params["charset"]

	require.Equal(r.testing, mediaType, "application/json")
	require.Contains(r.testing, []string{"utf-8", ""}, strings.ToLower(charset))

	return r
}

func (r *Response) ExpectList(values ...interface{}) *Response {
	actual := r.getList()

	if actual == nil {
		actual = make([]interface{}, 0, 0)
	}

	if values == nil {
		values = make([]interface{}, 0, 0)
	}

	require.Equal(r.testing, actual, values)

	return r
}

func (r *Response) ExpectListAnyOrder(values ...interface{}) *Response {
	actual := r.getList()

	require.Equal(r.testing, len(actual), len(values))

	if actual == nil {
		actual = make([]interface{}, 0, 0)
	}

	sorted := make([]interface{}, len(actual), len(actual))

	for _, v := range values {
		for n, a := range actual {
			if reflect.DeepEqual(v, a) {
				sorted[n] = v
				break
			}
		}
	}

	require.Equal(r.testing, actual, sorted)

	return r
}

func (r *Response) ExpectMap(value map[string]interface{}) *Response {
	actual := r.getMap()

	require.Equal(r.testing, actual, value)

	return r
}

func (r *Response) ExpectKey(key string) *Response {
	actual := r.getMap()

	require.Contains(r.testing, actual, key)

	return r
}

func (r *Response) getString() string {
	content, err := ioutil.ReadAll(r.resp.Body)
	if err != nil {
		require.FailNow(r.testing, err.Error())
	}

	return string(content)
}

func (r *Response) getList() []interface{} {
	r.ExpectJSON()

	content, err := ioutil.ReadAll(r.resp.Body)
	if err != nil {
		require.FailNow(r.testing, err.Error())
	}

	var result []interface{}
	if err := json.Unmarshal(content, &result); err != nil {
		require.FailNow(r.testing, err.Error())
	}

	return result
}

func (r *Response) getMap() map[string]interface{} {
	r.ExpectJSON()

	content, err := ioutil.ReadAll(r.resp.Body)
	if err != nil {
		require.FailNow(r.testing, err.Error())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(content, &result); err != nil {
		require.FailNow(r.testing, err.Error())
	}

	return result
}
