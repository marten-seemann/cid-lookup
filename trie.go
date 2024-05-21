package cidlookup

import (
	"errors"
)

const maxCIDLen = 20

var errCollision = errors.New("collisions")
var errNotFound = errors.New("not found")

type cidTrie struct {
	leaves [256]*cidTrie
	end    bool
}

func (t *cidTrie) Add(cid []byte) error {
	if len(cid) > maxCIDLen {
		return errors.New("CID is too long")
	}
	trie := t
	for i, b := range cid {
		nextLevel := trie.leaves[b]
		if nextLevel == nil {
			nextLevel = &cidTrie{}
			trie.leaves[b] = nextLevel
		} else if nextLevel.end {
			return errCollision
		}
		if i == len(cid)-1 {
			for _, l := range nextLevel.leaves {
				if l != nil {
					return errCollision
				}
			}
			nextLevel.end = true
		}
		trie = nextLevel
	}
	return nil
}

func (t *cidTrie) Lookup(data []byte) ([]byte, error) {
	if len(data) > maxCIDLen {
		return nil, errors.New("CID is too long")
	}
	if t.leaves[data[0]] == nil {
		return nil, errNotFound
	}
	nextLevel := t
	for i, b := range data {
		nextLevel = nextLevel.leaves[b]
		if nextLevel.end {
			return data[:i+1], nil
		}
	}
	return nil, errNotFound
}
