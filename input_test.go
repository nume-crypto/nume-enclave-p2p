package main

import "testing"

func TestGetInput(t *testing.T) {
	_, _, err := GetData("./test_data")
	if err != nil {
		t.Errorf("Error in getting input " + err.Error())
		return
	}
}
