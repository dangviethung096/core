package core

const NEAREST_NODE_COUNT = 10

func NewSearchTree[T any]() *SearchTree[T] {
	return &SearchTree[T]{
		root: &SearchTreeNode[T]{
			children: make(map[rune]*SearchTreeNode[T]),
		},
	}
}

type SearchTree[T any] struct {
	root *SearchTreeNode[T]
}

type SearchTreeNode[T any] struct {
	char      rune
	children  map[rune]*SearchTreeNode[T]
	value     T
	haveValue bool
}

func (t *SearchTree[T]) Insert(key string, value T) {
	node := t.root
	for _, char := range key {
		if node.children[char] == nil {
			node.children[char] = &SearchTreeNode[T]{
				char:     char,
				children: make(map[rune]*SearchTreeNode[T]),
			}
		}
		node = node.children[char]
	}
	node.value = value
	node.haveValue = true
}

func (t *SearchTree[T]) Search(key string) []T {
	node := t.root
	for _, char := range key {
		if node.children[char] == nil {
			break
		}
		node = node.children[char]
	}

	if len(node.children) == 0 {
		return []T{node.value}
	}

	result := []T{}
	queue := []*SearchTreeNode[T]{
		node,
	}
	// Find 10 nearest left
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		if node.haveValue {
			result = append(result, node.value)
			if len(result) == NEAREST_NODE_COUNT {
				break
			}
		}

		for _, child := range node.children {
			queue = append(queue, child)
		}
	}

	return result
}

func (t *SearchTree[T]) Remove(key string) bool {
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

func (t *SearchTree[T]) PrintTree() {
	type queueElement struct {
		node   *SearchTreeNode[T]
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
