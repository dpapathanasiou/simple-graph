package simplegraph

import (
	"bytes"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

type EdgeSet struct {
	set map[EdgeData]bool
}

func NewEdgeSet() *EdgeSet {
	return &EdgeSet{make(map[EdgeData]bool)}
}

func (set *EdgeSet) Add(edge EdgeData) bool {
	_, found := set.set[edge]
	set.set[edge] = true
	return !found
}

func (set *EdgeSet) Contains(edge EdgeData) bool {
	_, found := set.set[edge]
	return found
}

func Visualize(path []string, database ...string) string {
	gv := graphviz.New()
	graph, err := gv.Graph()
	evaluate(err)
	defer func() {
		err := graph.Close()
		evaluate(err)
		gv.Close()
	}()

	nodes := make(map[string]*cgraph.Node)
	plotted := NewEdgeSet()
	for _, identifier := range path {
		var node *cgraph.Node
		body, err := FindNode(identifier, database...)
		evaluate(err)
		node, err = graph.CreateNode(identifier)
		evaluate(err)
		node.SetLabel(body)
		nodes[identifier] = node

		edges, err := Connections(identifier, database...)
		evaluate(err)
		for _, edge := range edges {
			if !plotted.Contains(edge) {
				plotted.Add(edge)
				_, exists := nodes[edge.Target]
				if !exists {
					body, err = FindNode(edge.Target, database...)
					evaluate(err)
					target, err := graph.CreateNode(edge.Target)
					evaluate(err)
					target.SetLabel(body)
					nodes[edge.Target] = target
				}
				target := nodes[edge.Target]
				if node != target {
					e, err := graph.CreateEdge("", node, target)
					evaluate(err)
					if len(edge.Label) > 0 && edge.Label != `{}` {
						e.SetLabel(edge.Label)
					}
				}
			}
		}
	}

	/*
		Could also render the image directly:

		gv.RenderFilename(graph, graphviz.PNG, "/path/to/graph.png")

		instead of returning the dot file string as this function does:
	*/

	var buf bytes.Buffer
	err = gv.Render(graph, "dot", &buf)
	evaluate(err)
	return buf.String()
}

func VisualizeBodies(path []GraphData, database ...string) string {
	gv := graphviz.New()
	graph, err := gv.Graph()
	evaluate(err)
	defer func() {
		err := graph.Close()
		evaluate(err)
		gv.Close()
	}()

	nodes := make(map[string]*cgraph.Node)

	for _, object := range path {
		var node *cgraph.Node
		var exists bool
		var err error
		if object.Node.Identifier != nil {
			id := object.Node.Identifier.(string)
			_, exists = nodes[id]
			if !exists {
				node, err = graph.CreateNode(id)
				evaluate(err)
				node.SetLabel(object.Node.Body.(string))
				nodes[id] = node
			}
		}
	}

	for _, object := range path {
		if object.Node.Identifier == nil {
			source, sourceExists := nodes[object.Edge.Source]
			target, targetExists := nodes[object.Edge.Target]
			if sourceExists && targetExists {
				edge, err := graph.CreateEdge("", source, target)
				evaluate(err)
				if object.Edge.Label != `{}` {
					edge.SetLabel(object.Edge.Label)
				}
			}
		}
	}

	/*
		Could also render the image directly:

		gv.RenderFilename(graph, graphviz.PNG, "/path/to/graph.png")

		instead of returning the dot file string as this function does:
	*/

	var buf bytes.Buffer
	err = gv.Render(graph, "dot", &buf)
	evaluate(err)
	return buf.String()
}
