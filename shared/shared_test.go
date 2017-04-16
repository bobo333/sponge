package shared_test

import "fmt"
import "os"
import shared "github.com/bobo333/sponge/shared"
import "testing"

func TestGetEnvVar(t *testing.T) {
	// nonexistent should return an error
	val, err := shared.GetEnvVar("DOES_NOT_EXIST")
	if err == nil {
		t.Error("Expected err to not be nil for nonexistent env var")
	}
	if val != "" {
		t.Error("Expected empty value for nonexistent env var")
	}

	// env var that exists
	my_env_var := "MY_ENV_VAR"
	my_value := "my_value"
	os.Setenv(my_env_var, my_value)
	val, err = shared.GetEnvVar(my_env_var)
	if err != nil {
		t.Error("Expected nil error if env var exists")
	}
	if val != my_value {
		t.Errorf("Expected env var to be %s, got %s", my_value, val)
	}
}

func TestCreateRawTest(t *testing.T) {
	// no sections
	output := shared.CreateRawText(make([]shared.OutputSection, 0))
	if output != "" {
		t.Errorf("Expected empty output for no sections, got %s", output)
	}

	// one section
	sectionName := "section name"
	itemName := "item name"
	itemUrl := "a url"
	item := shared.StandardizedItem{Title: itemName, Url: itemUrl}
	outputSection := shared.OutputSection{Name: sectionName, Items: []shared.StandardizedItem{item}}

	expectedOutput := fmt.Sprintf(`

=====================================
%s
=====================================

Title: %s
Url: %s

`, sectionName, itemName, itemUrl)
	output = shared.CreateRawText([]shared.OutputSection{outputSection})
	if output != expectedOutput {
		t.Errorf("expected %s got %s", expectedOutput, output)
	}

	// multiple sections
	section1Name := "section 1 name"
	item1Name := "item 1 name"
	item1Url := "1 url"
	item1 := shared.StandardizedItem{Title: item1Name, Url: item1Url}
	output1Section := shared.OutputSection{Name: section1Name, Items: []shared.StandardizedItem{item1}}

	section2Name := "section 2 name"
	item2Name := "item 2 name"
	item2Url := "2 url"
	item2 := shared.StandardizedItem{Title: item2Name, Url: item2Url}
	item3Name := "item 3 name"
	item3Url := "3 url"
	item3Comments := "comments url"
	item3 := shared.StandardizedItem{Title: item3Name, Url: item3Url, Comments: item3Comments}
	output2Section := shared.OutputSection{Name: section2Name, Items: []shared.StandardizedItem{item2, item3}}

	expectedOutput = fmt.Sprintf(`

=====================================
%s
=====================================

Title: %s
Url: %s



=====================================
%s
=====================================

Title: %s
Url: %s

Title: %s
Url: %s
Comments: %s

`, section1Name, item1Name, item1Url, section2Name, item2Name, item2Url, item3Name, item3Url, item3Comments)

	output = shared.CreateRawText([]shared.OutputSection{output1Section, output2Section})
	if output != expectedOutput {
		t.Errorf("expected %s got %s", expectedOutput, output)
	}
}
