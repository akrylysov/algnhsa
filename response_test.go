package algnhsa

import "testing"

func TestBinaryCase(t *testing.T) {
	if binaryCase("ab", 0) != "ab" {
		t.Errorf("binaryCase (ab,0) expected %s observed %s", "ab", binaryCase("ab", 0))
	}
	if binaryCase("ab", 1) != "Ab" {
		t.Errorf("binaryCase (ab,1) expected %s observed %s", "Ab", binaryCase("ab", 1))
	}
	if binaryCase("ab", 2) != "aB" {
		t.Errorf("binaryCase (ab,2) expected %s observed %s", "aB", binaryCase("ab", 2))
	}
	if binaryCase("ab", 3) != "AB" {
		t.Errorf("binaryCase (ab,3) expected %s observed %s", "AB", binaryCase("ab", 3))
	}
	if binaryCase("a----b", 3) != "A----B" {
		t.Errorf("binaryCase (a----b,3) expected %s observed %s", "A----B", binaryCase("a----b", 3))
	}
}
