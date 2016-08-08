package kk

type INeuron interface {
	Name() string
	Send(message *Message, from INeuron)
	Address() string
	Break()
}

type Neuron struct {
	name    string
	address string
}

func (n *Neuron) Name() string {
	return n.name
}

func (n *Neuron) Address() string {
	return n.address
}
