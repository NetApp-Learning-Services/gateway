package ontap_test

import (
	"gateway/internal/controller/ontap"
	"testing"
)

func TestParseUUIDWithLetter(t *testing.T) {
	input := "POST /api/svm/svms/a1278052-4bd3-11ed-9beb-005056b09977"
	char := "/"
	expectedOutput := "a1278052-4bd3-11ed-9beb-005056b09977"

	actualOutput, actualError := ontap.ParseUUID(input, char)

	if actualError != nil {
		t.Errorf("Expected no error, but found %v", actualError)
	}

	if actualOutput != expectedOutput {
		t.Errorf("Expected out to be %s, but found %s", expectedOutput, actualOutput)
	}

}

func TestParseUUIDWithNumber(t *testing.T) {
	input := "POST /api/svm/svms/31278052-4bd3-11ed-9beb-005056b09977"
	char := "/"
	expectedOutput := "31278052-4bd3-11ed-9beb-005056b09977"

	actualOutput, actualError := ontap.ParseUUID(input, char)

	if actualError != nil {
		t.Errorf("Expected no error, but found %v", actualError)
	}

	if actualOutput != expectedOutput {
		t.Errorf("Expected out to be %s, but found %s", expectedOutput, actualOutput)
	}

}
