package simplegraph

import (
	"fmt"
	"os"
	"testing"
)

func TestVisualization(t *testing.T) {
	file := "testdb.sqlite3"
	Initialize(file)
	defer os.Remove(file)

	/*
		Test using the graph from:
		https://commons.wikimedia.org/wiki/File:Graph_with_Chordless_and_Chorded_Cycles.svg
	*/
	nodes := [][]byte{}
	ids := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L"}
	l := len(ids)
	for i := 0; i < l; i++ {
		nodes = append(nodes, []byte(fmt.Sprintf("{\"id\":%q}", ids[i])))
	}

	count, err := AddNodes(ids, nodes, file)
	if count != 1 && err != nil {
		t.Errorf("AddNode() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	sources := []string{"A", "A", "B", "B", "B", "C", "C", "C", "C", "D", "D", "D", "E", "E", "G", "G", "G", "G", "H", "H", "I", "I", "I", "J", "J", "K", "K", "K", "K", "L", "L", "L"}
	targets := []string{"B", "F", "A", "C", "G", "B", "D", "G", "L", "C", "E", "K", "D", "F", "B", "C", "H", "L", "G", "I", "H", "J", "K", "I", "K", "D", "I", "J", "L", "C", "G", "K"}

	count, err = BulkConnectNodes(sources, targets, file)
	if count != 1 && err != nil {
		t.Errorf("BulkConnectNodes() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	dot := Visualize([]string{"A", "B", "G", "H", "I"}, file)
	expected := 1740
	if len(dot) != expected {
		t.Errorf("Visualize() produced string of len %d but expected %d", len(dot), expected)
	}
}
