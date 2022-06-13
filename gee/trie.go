package gee

import "strings"

/*
	目前支持：:name和*filepath两种模式
*/
type node struct {
	pattern  string  // 待匹配路由，例如 /p/:lang, 非叶子节点的pattern都为nil
	part     string  // 路由中的一部分，例如 :lang
	children []*node // 孩子节点
	isWild   bool    //是否精确匹配，part 含有 : 或 * 时为true
}

// 第一个匹配成功的节点，用于插入
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 所有匹配成功的节点，用于查找
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		// 如果part是匹配的，或者part中含有:或者*,则匹配成功
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

// 向前缀树中插入节点
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		return
	}
	// 获取要插入的part值
	part := parts[height]
	// 查找是否有匹配的子节点
	child := n.matchChild(part)
	// 如果没有匹配的子节点，则创建一个
	if child == nil {
		child = &node{
			part:   part,
			isWild: part[0] == ':' || part[0] == '*',
		}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

// 查找匹配的节点
func (n *node) search(parts []string, height int) *node {
	// 如果part的前缀为*，直接匹配成功
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}
	part := parts[height]
	children := n.matchChildren(part)
	for _, child := range children {
		// 从当前子节点递归查找
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}
	return nil
}
