package algnhsa

import "testing"

func TestBinaryCase(t *testing.T) {
	testCases := []struct {
		in  string
		n   int
		out string
	}{
		{
			in:  "ab",
			n:   0,
			out: "ab",
		},
		{
			in:  "ab",
			n:   1,
			out: "Ab",
		},
		{
			in:  "ab",
			n:   2,
			out: "aB",
		},
		{
			in:  "ab",
			n:   3,
			out: "AB",
		},
		{
			in:  "a----b",
			n:   3,
			out: "A----B",
		},
	}
	for _, testCase := range testCases {
		if actualOut := binaryCase(testCase.in, testCase.n); actualOut != testCase.out {
			t.Errorf("binaryCase(%s,%d) expected %s observed %s", testCase.in, testCase.n, testCase.out, actualOut)
		}
	}
}

func BenchmarkBinaryCase0(t *testing.B) {
	for i := 0; i < t.N; i++ {
		binaryCase("X-Content-Type-Options", 0)
	}
}

func BenchmarkBinaryCase1(t *testing.B) {
	for i := 0; i < t.N; i++ {
		binaryCase("X-Content-Type-Options", 1)
	}
}

func BenchmarkBinaryCase8(t *testing.B) {
	for i := 0; i < t.N; i++ {
		binaryCase("X-Content-Type-Options", 8)
	}
}
