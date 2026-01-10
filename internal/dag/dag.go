package dag

type DAG struct {
	Nodes map[string]*Node
	Root  *Node
}

func New() *DAG {
	return &DAG{
		Nodes: make(map[string]*Node),
	}
}
