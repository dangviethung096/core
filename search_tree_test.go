package core

import "testing"

func TestSearchTree(t *testing.T) {
	type test struct {
		name  string
		value string
	}
	tree := NewSearchTree[test]()

	tree.Insert("hello", test{name: "hello", value: "xin chào"})
	tree.Insert("how", test{name: "how", value: "thế nào"})
	tree.Insert("hi", test{name: "hi", value: "xin chào"})
	tree.Insert("bye", test{name: "bye", value: "tạm biệt"})

	result := tree.Search("h")

	if len(result) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result))
	}

	for i, item := range result {
		t.Logf("Result[%d]: %v", i, item)
	}
}

func TestSearchTree_MoreThan10Element(t *testing.T) {
	type test struct {
		name  string
		value string
	}
	tree := NewSearchTree[test]()

	tree.Insert("hello", test{name: "hello", value: "xin chào"})
	tree.Insert("how", test{name: "how", value: "thế nào"})
	tree.Insert("hi", test{name: "hi", value: "xin chào"})
	tree.Insert("hell", test{name: "hell", value: "địa ngục"})
	tree.Insert("he", test{name: "he", value: "anh ấy"})
	tree.Insert("hey", test{name: "hey", value: "hey"})
	tree.Insert("hang", test{name: "hang", value: "hằng"})
	tree.Insert("hong", test{name: "hong", value: "hồng"})
	tree.Insert("hurry", test{name: "hurry", value: "nhanh"})
	tree.Insert("hurt", test{name: "hurt", value: "đau"})
	tree.Insert("hung", test{name: "hung", value: "hưng"})

	tree.Insert("bye", test{name: "bye", value: "tạm biệt"})

	result := tree.Search("h")

	if len(result) != 10 {
		t.Errorf("Expected 10 results, got %d", len(result))
	}

	for i, item := range result {
		t.Logf("Result[%d]: %v", i, item)
	}
}

func TestPrintTree(t *testing.T) {
	type test struct {
		name  string
		value string
	}
	tree := NewSearchTree[test]()

	tree.Insert("hello", test{name: "hello", value: "xin chào"})
	tree.Insert("how", test{name: "how", value: "thế nào"})
	tree.Insert("hi", test{name: "hi", value: "xin chào"})
	tree.Insert("hell", test{name: "hell", value: "địa ngục"})
	tree.Insert("he", test{name: "he", value: "anh ấy"})
	tree.Insert("hey", test{name: "hey", value: "hey"})
	tree.Insert("hang", test{name: "hang", value: "hằng"})
	tree.Insert("hong", test{name: "hong", value: "hồng"})
	tree.Insert("hurry", test{name: "hurry", value: "nhanh"})
	tree.Insert("hurt", test{name: "hurt", value: "đau"})
	tree.Insert("hung", test{name: "hung", value: "hưng"})
	tree.Insert("bye", test{name: "bye", value: "tạm biệt"})

	tree.PrintTree()
}
