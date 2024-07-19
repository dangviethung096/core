package core

type SearchTree[T any] interface {
	Insert(key string, value T)
	Search(key string) []T
	SearchNearestNValue(key string, n int) []T
	SearchSpecificValue(key string) (T, bool)
	Remove(key string) bool
	PrintTree()
}

func NewSearchTree[T any]() SearchTree[T] {
	return &searchTree[T]{
		root: &searchTreeNode[T]{
			children:  make(map[rune]*searchTreeNode[T]),
			haveValue: false,
		},
	}
}

type searchTree[T any] struct {
	root *searchTreeNode[T]
}

type searchTreeNode[T any] struct {
	char      rune
	children  map[rune]*searchTreeNode[T]
	value     T
	haveValue bool
}

func (t *searchTree[T]) Insert(key string, value T) {
	node := t.root
	for _, char := range key {
		if node.children[char] == nil {
			node.children[char] = &searchTreeNode[T]{
				char:     char,
				children: make(map[rune]*searchTreeNode[T]),
			}
		}
		node = node.children[char]
	}
	node.value = value
	node.haveValue = true
}

func (t *searchTree[T]) Search(key string) []T {
	node := t.root
	for _, char := range key {
		if node.children[char] == nil {
			return []T{}
		}
		node = node.children[char]
	}

	result := []T{}
	if node.haveValue {
		result = append(result, node.value)
	}

	// Find all in node
	queue := []*searchTreeNode[T]{}

	for _, node := range node.children {
		queue = append(queue, node)
	}

	// Find all value from node
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		if node.haveValue {
			result = append(result, node.value)
		}

		for _, child := range node.children {
			queue = append(queue, child)
		}
	}

	return result
}

func (t *searchTree[T]) SearchNearestNValue(key string, n int) []T {
	if n <= 0 {
		return nil
	}

	node := t.root
	for _, char := range key {
		if node.children[char] == nil {
			return []T{}
		}
		node = node.children[char]
	}

	result := []T{}
	if node.haveValue {
		result = append(result, node.value)
		if n == 1 {
			return result
		}
	}

	// Find all in node
	queue := []*searchTreeNode[T]{}

	for _, node := range node.children {
		queue = append(queue, node)
	}

	// Find all value from node
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		if node.haveValue {
			result = append(result, node.value)
			if len(result) == n {
				break
			}
		}

		for _, child := range node.children {
			queue = append(queue, child)
		}
	}

	return result
}

func (t *searchTree[T]) SearchSpecificValue(key string) (T, bool) {
	var emptyValue T
	node := t.root
	for _, char := range key {
		if node.children[char] == nil {
			return emptyValue, false
		}
		node = node.children[char]
	}

	if node.haveValue {
		return node.value, true
	}

	return emptyValue, false
}

func (t *searchTree[T]) Remove(key string) bool {
	node := t.root
	for _, char := range key {
		if node.children[char] == nil {
			return false
		}
		node = node.children[char]
	}
	node.haveValue = false
	return true
}

func (t *searchTree[T]) PrintTree() {
	type queueElement struct {
		node   *searchTreeNode[T]
		prefix string
	}
	queue := []queueElement{
		{node: t.root, prefix: ""},
	}

	for len(queue) > 0 {
		element := queue[0]
		queue = queue[1:]
		LogInfo("Browse: %s: value = %v", element.prefix+string(element.node.char), element.node.value)
		if element.node.haveValue {
			LogInfo(element.prefix + string(element.node.char))

		}

		for _, child := range element.node.children {
			queue = append(queue, queueElement{
				node:   child,
				prefix: element.prefix + string(element.node.char),
			})
		}
	}
}
