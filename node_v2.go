package gotaskflow


type GGraph struct {
	name string
	state      kNodeState
	successors []*GGraph
	dependents []*GGraph
}