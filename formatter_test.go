package httpexpect

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type typedNil int

func (*typedNil) String() string {
	return ""
}

func TestFormatValues(t *testing.T) {
	checkAll := func(t *testing.T, fn func(interface{}) string) {
		var tnil *typedNil
		var tnilPtr fmt.Stringer = tnil

		assert.Nil(t, tnilPtr)
		assert.NotEqual(t, nil, tnilPtr)

		check := func(s string) {
			t.Logf("\n%s", s)
			assert.NotEmpty(t, s)
		}

		check(fn(nil))
		check(fn(tnil))
		check(fn(tnilPtr))
		check(fn(123))
		check(fn("hello"))
		check(fn(time.Second))
		check(fn(time.Unix(0, 0)))
		check(fn([]interface{}{1, 2}))
		check(fn(map[string]string{"a": "b"}))
		check(fn(make(chan int)))
		check(fn(AssertionRange{1, 2}))
		check(fn(&AssertionRange{1, 2}))
		check(fn(AssertionRange{"a", "b"}))
		check(fn(AssertionList([]interface{}{1, 2})))
	}

	t.Run("formatTypes", func(t *testing.T) {
		checkAll(t, formatTyped)
	})

	t.Run("formatValue", func(t *testing.T) {
		checkAll(t, formatValue)
	})

	t.Run("formatBareString", func(t *testing.T) {
		checkAll(t, formatBareString)
	})

	t.Run("formatRange", func(t *testing.T) {
		checkAll(t, func(v interface{}) string {
			return strings.Join(formatRange(v), "")
		})
	})

	t.Run("formatList", func(t *testing.T) {
		checkAll(t, func(v interface{}) string {
			return strings.Join(formatList(v), "")
		})
	})
}

func TestFormatDiff(t *testing.T) {
	checkOK := func(a, b interface{}) {
		s, ok := formatDiff(a, b)
		assert.True(t, ok)
		assert.NotEqual(t, "", s)
	}

	checkNotOK := func(a, b interface{}) {
		s, ok := formatDiff(a, b)
		assert.False(t, ok)
		assert.Equal(t, "", s)
	}

	checkNotOK(map[string]interface{}{}, []interface{}{})
	checkNotOK([]interface{}{}, map[string]interface{}{})
	checkNotOK("foo", "bar")
	checkNotOK(func() {}, func() {})

	checkNotOK(map[string]interface{}{}, map[string]interface{}{})
	checkNotOK([]interface{}{}, []interface{}{})

	checkOK(map[string]interface{}{"a": 1}, map[string]interface{}{})
	checkOK([]interface{}{"a"}, []interface{}{})
}

func TestFormatDataFailureActual(t *testing.T) {
	df := &DefaultFormatter{}
	ctx := &AssertionContext{}

	t.Run("AssertType nil", func(t *testing.T) {
		fl := &AssertionFailure{
			Type: AssertType,
			Actual: &AssertionValue{
				Value: nil,
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveActual)
		assert.Equal(t, "<nil>(<nil>)", fd.Actual)
	})

	t.Run("AssertType int", func(t *testing.T) {
		fl := &AssertionFailure{
			Type: AssertType,
			Actual: &AssertionValue{
				Value: int(1_000_000),
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveActual)
		assert.Equal(t, "int(1000000)", fd.Actual)

	})

	t.Run("AssertType float32", func(t *testing.T) {
		fl := &AssertionFailure{
			Type: AssertType,
			Actual: &AssertionValue{
				Value: float32(1_000_000),
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveActual)
		assert.Equal(t, "float32(1e+06)", fd.Actual)
	})

	t.Run("AssertType float64", func(t *testing.T) {
		fl := &AssertionFailure{
			Type: AssertType,
			Actual: &AssertionValue{
				Value: float64(1_000_000),
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveActual)
		assert.Equal(t, "float64(1e+06)", fd.Actual)
	})

	t.Run("AssertType string", func(t *testing.T) {
		fl := &AssertionFailure{
			Type: AssertType,
			Actual: &AssertionValue{
				Value: "value string",
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveActual)
		assert.Equal(t, "string(\"value string\")", fd.Actual)
	})

	t.Run("AssertType object", func(t *testing.T) {
		obj := struct{ Name string }{"testName"}
		fl := &AssertionFailure{
			Type: AssertType,
			Actual: &AssertionValue{
				Value: obj,
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveActual)
		assert.Equal(
			t,
			"struct { Name string }(struct { Name string }{Name:\"testName\"})",
			fd.Actual,
		)
	})
}

func TestFormatDataFailureExpected(t *testing.T) {
	df := &DefaultFormatter{}
	ctx := &AssertionContext{}

	t.Run("AssertInRange", func(t *testing.T) {
		t.Run("nil", func(t *testing.T) {
			fl := &AssertionFailure{
				Type: AssertInRange,
				Expected: &AssertionValue{
					Value: AssertionRange{
						Min: nil,
						Max: nil,
					},
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.True(t, fd.HaveExpected)
			assert.Equal(t, kindRange, fd.ExpectedKind)
			assert.Equal(t, []string{"<nil>", "<nil>"}, fd.Expected)
		})

		t.Run("int", func(t *testing.T) {
			fl := &AssertionFailure{
				Type: AssertInRange,
				Expected: &AssertionValue{
					Value: AssertionRange{
						Min: int(1_000_000),
						Max: int(2_000_000),
					},
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.True(t, fd.HaveExpected)
			assert.Equal(t, kindRange, fd.ExpectedKind)
			assert.Equal(t, []string{"[1000000; 2000000]"}, fd.Expected)
		})

		t.Run("float32", func(t *testing.T) {
			fl := &AssertionFailure{
				Type: AssertInRange,
				Expected: &AssertionValue{
					Value: AssertionRange{
						Min: float32(1_000_000),
						Max: float32(2_000_000),
					},
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.True(t, fd.HaveExpected)
			assert.Equal(t, kindRange, fd.ExpectedKind)
			assert.Equal(t, []string{"[1e+06; 2e+06]"}, fd.Expected)
		})

		t.Run("float64", func(t *testing.T) {
			fl := &AssertionFailure{
				Type: AssertInRange,
				Expected: &AssertionValue{
					Value: AssertionRange{
						Min: float64(1_000_000),
						Max: float64(2_000_000),
					},
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.True(t, fd.HaveExpected)
			assert.Equal(t, kindRange, fd.ExpectedKind)
			assert.Equal(t, []string{"[1e+06; 2e+06]"}, fd.Expected)
		})

		t.Run("string", func(t *testing.T) {
			fl := &AssertionFailure{
				Type: AssertInRange,
				Expected: &AssertionValue{
					Value: AssertionRange{
						Min: "string 1",
						Max: "string 2",
					},
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.True(t, fd.HaveExpected)
			assert.Equal(t, kindRange, fd.ExpectedKind)
			assert.Equal(t, []string{"string 1", "string 2"}, fd.Expected)
		})

		t.Run("object", func(t *testing.T) {
			obj1 := struct{ Name string }{"testName1"}
			obj2 := struct{ Name string }{"testName2"}
			fl := &AssertionFailure{
				Type: AssertInRange,
				Expected: &AssertionValue{
					Value: AssertionRange{
						Min: obj1,
						Max: obj2,
					},
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.True(t, fd.HaveExpected)
			assert.Equal(t, kindRange, fd.ExpectedKind)
			assert.Equal(t, []string{"{testName1}", "{testName2}"}, fd.Expected)
		})

	})

	t.Run("AssertMatchPath", func(t *testing.T) {
		t.Run("nil", func(t *testing.T) {
			fl := &AssertionFailure{
				Type: AssertMatchPath,
				Expected: &AssertionValue{
					Value: nil,
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.True(t, fd.HaveExpected)
			assert.Equal(t, kindPath, fd.ExpectedKind)
			assert.Equal(t, []string{"nil"}, fd.Expected)
		})

		t.Run("int", func(t *testing.T) {
			fl := &AssertionFailure{
				Type: AssertMatchPath,
				Expected: &AssertionValue{
					Value: int(1_000_000),
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.True(t, fd.HaveExpected)
			assert.Equal(t, kindPath, fd.ExpectedKind)
			assert.Equal(t, []string{"1000000"}, fd.Expected)
		})

		t.Run("float32", func(t *testing.T) {
			fl := &AssertionFailure{
				Type: AssertMatchPath,
				Expected: &AssertionValue{
					Value: float32(1_000_000),
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.True(t, fd.HaveExpected)
			assert.Equal(t, kindPath, fd.ExpectedKind)
			assert.Equal(t, []string{"1e+06"}, fd.Expected)
		})

		t.Run("float64", func(t *testing.T) {
			fl := &AssertionFailure{
				Type: AssertMatchPath,
				Expected: &AssertionValue{
					Value: float64(1_000_000),
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.True(t, fd.HaveExpected)
			assert.Equal(t, kindPath, fd.ExpectedKind)
			assert.Equal(t, []string{"1e+06"}, fd.Expected)
		})

		t.Run("string", func(t *testing.T) {
			fl := &AssertionFailure{
				Type: AssertMatchPath,
				Expected: &AssertionValue{
					Value: "match path string",
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.True(t, fd.HaveExpected)
			assert.Equal(t, kindPath, fd.ExpectedKind)
			assert.Equal(t, []string{"match path string"}, fd.Expected)
		})

		t.Run("object", func(t *testing.T) {
			obj := struct{ Name string }{"testName"}
			fl := &AssertionFailure{
				Type: AssertMatchPath,
				Expected: &AssertionValue{
					Value: obj,
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.True(t, fd.HaveExpected)
			assert.Equal(t, kindPath, fd.ExpectedKind)
			assert.Equal(t, []string{"{\n  \"Name\": \"testName\"\n}"}, fd.Expected)
		})

	})

	t.Run("AssertMatchFormat", func(t *testing.T) {
		t.Run("int", func(t *testing.T) {
			fl := &AssertionFailure{
				Type: AssertMatchFormat,
				Expected: &AssertionValue{
					Value: AssertionList{
						nil,
						nil,
					},
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.True(t, fd.HaveExpected)
			assert.Equal(t, kindFormatList, fd.ExpectedKind)
			assert.Equal(t, []string{"nil", "nil"}, fd.Expected)
		})

		t.Run("int", func(t *testing.T) {
			fl := &AssertionFailure{
				Type: AssertMatchFormat,
				Expected: &AssertionValue{
					Value: AssertionList{
						int(1_000_000),
						int(2_000_000),
					},
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.True(t, fd.HaveExpected)
			assert.Equal(t, kindFormatList, fd.ExpectedKind)
			assert.Equal(t, []string{"1000000", "2000000"}, fd.Expected)
		})

		t.Run("float32", func(t *testing.T) {
			fl := &AssertionFailure{
				Type: AssertMatchFormat,
				Expected: &AssertionValue{
					Value: AssertionList{
						float32(1_000_000),
						float32(2_000_000),
					},
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.True(t, fd.HaveExpected)
			assert.Equal(t, kindFormatList, fd.ExpectedKind)
			assert.Equal(t, []string{"1e+06", "2e+06"}, fd.Expected)
		})

		t.Run("float64", func(t *testing.T) {
			fl := &AssertionFailure{
				Type: AssertMatchFormat,
				Expected: &AssertionValue{
					Value: AssertionList{
						float64(1_000_000),
						float64(2_000_000),
					},
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.True(t, fd.HaveExpected)
			assert.Equal(t, kindFormatList, fd.ExpectedKind)
			assert.Equal(t, []string{"1e+06", "2e+06"}, fd.Expected)
		})

		t.Run("string", func(t *testing.T) {
			fl := &AssertionFailure{
				Type: AssertMatchFormat,
				Expected: &AssertionValue{
					Value: AssertionList{
						"string 1",
						"string 2",
					},
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.True(t, fd.HaveExpected)
			assert.Equal(t, kindFormatList, fd.ExpectedKind)
			assert.Equal(t, []string{"\"string 1\"", "\"string 2\""}, fd.Expected)
		})

		t.Run("object", func(t *testing.T) {
			obj1 := struct{ Name string }{"testName1"}
			obj2 := struct{ Name string }{"testName2"}
			fl := &AssertionFailure{
				Type: AssertMatchFormat,
				Expected: &AssertionValue{
					Value: AssertionList{
						obj1,
						obj2,
					},
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.True(t, fd.HaveExpected)
			assert.Equal(t, kindFormatList, fd.ExpectedKind)
			assert.Equal(
				t,
				[]string{"{\n  \"Name\": \"testName1\"\n}", "{\n  \"Name\": \"testName2\"\n}"},
				fd.Expected,
			)
		})
	})
}

func TestFormatDataFailureReference(t *testing.T) {
	df := &DefaultFormatter{}
	ctx := &AssertionContext{}

	t.Run("nil", func(t *testing.T) {
		fl := &AssertionFailure{
			Reference: &AssertionValue{
				Value: nil,
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveReference)
		assert.Equal(t, "nil", fd.Reference)
	})

	t.Run("int", func(t *testing.T) {
		fl := &AssertionFailure{
			Reference: &AssertionValue{
				Value: int(1_000_000),
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveReference)
		assert.Equal(t, "1000000", fd.Reference)
	})

	t.Run("float32", func(t *testing.T) {
		fl := &AssertionFailure{
			Reference: &AssertionValue{
				Value: float32(1_000_000),
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveReference)
		assert.Equal(t, "1e+06", fd.Reference)

	})

	t.Run("float64", func(t *testing.T) {
		fl := &AssertionFailure{
			Reference: &AssertionValue{
				Value: float64(1_000_000),
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveReference)
		assert.Equal(t, "1e+06", fd.Reference)
	})

	t.Run("string", func(t *testing.T) {
		fl := &AssertionFailure{
			Reference: &AssertionValue{
				Value: "reference string",
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveReference)
		assert.Equal(t, "\"reference string\"", fd.Reference)
	})

	t.Run("object", func(t *testing.T) {
		obj := struct{ Name string }{"testName"}
		fl := &AssertionFailure{
			Reference: &AssertionValue{
				Value: obj,
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveReference)
		assert.Equal(t, "{\n  \"Name\": \"testName\"\n}", fd.Reference)
	})
}

func TestFormatDataFailureDelta(t *testing.T) {
	df := &DefaultFormatter{}
	ctx := &AssertionContext{}

	t.Run("nil", func(t *testing.T) {
		fl := &AssertionFailure{
			Delta: &AssertionValue{
				Value: nil,
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveDelta)
		assert.Equal(t, "nil", fd.Delta)
	})

	t.Run("int", func(t *testing.T) {
		fl := &AssertionFailure{
			Delta: &AssertionValue{
				Value: int(1_000_000),
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveDelta)
		assert.Equal(t, "1000000", fd.Delta)
	})

	t.Run("float32", func(t *testing.T) {
		fl := &AssertionFailure{
			Delta: &AssertionValue{
				Value: float32(1_000_000),
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveDelta)
		assert.Equal(t, "1e+06", fd.Delta)

	})

	t.Run("float64", func(t *testing.T) {
		fl := &AssertionFailure{
			Delta: &AssertionValue{
				Value: float64(1_000_000),
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveDelta)
		assert.Equal(t, "1e+06", fd.Delta)
	})

	t.Run("string", func(t *testing.T) {
		fl := &AssertionFailure{
			Delta: &AssertionValue{
				Value: "delta string",
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveDelta)
		assert.Equal(t, `"delta string"`, fd.Delta)
	})

	t.Run("string", func(t *testing.T) {
		obj := struct{ Name string }{"testName"}
		fl := &AssertionFailure{
			Delta: &AssertionValue{
				Value: obj,
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveDelta)
		assert.Equal(t, "{\n  \"Name\": \"testName\"\n}", fd.Delta)
	})
}
