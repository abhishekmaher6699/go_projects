package main

import (
	"fmt"
	"hash/fnv"
	"slices"
	"sort"
)

type HashRing struct {
	nodes map[uint32]string
	keys  []uint32
}

func NewHashRing() *HashRing {
	return &HashRing{
		nodes: make(map[uint32]string),
	}
}

const virtualNodes = 100

func (h *HashRing) AddNode(node string) {

	for i := range virtualNodes {

		virtualNode := fmt.Sprintf("%s-%d", node, i)
		hash := hashKey(virtualNode)

		h.nodes[hash] = node
		h.keys = append(h.keys, hash)

	}

	slices.Sort(h.keys)
}

func (h *HashRing) GetNodes(key string, count int) []string {

	if len(h.keys) == 0 || count <= 0 {
		return nil
	}

	hash := hashKey(key)

	index := sort.Search(
		len(h.keys),
		func(i int) bool {
			return h.keys[i] >= hash
		},
	)

	if index == len(h.keys) {
		index = 0
	}

	result := make([]string, 0, count)
	seen := make(map[string]bool)

	for len(result) < count && len(seen) < len(h.nodes) {

		node := h.nodes[h.keys[index]]

		if !seen[node] {
			result = append(result, node)
			seen[node] = true
		}

		index = (index + 1) % len(h.keys)
	}

	return result
}

func hashKey(key string) uint32 {
	hash := fnv.New32a()
	hash.Write([]byte(key))

	return hash.Sum32()
}
