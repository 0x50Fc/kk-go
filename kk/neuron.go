package kk

type INeuron interface {
	Name() string
	Send(message *Message, from INeuron)
	Address() string
	Break()
	Get(key string) interface{}
	Set(key string, value interface{})
	Remove(key string)
}

type Neuron struct {
	name    string
	address string
	values  map[string]interface{}
}

func (n *Neuron) Name() string {
	return n.name
}

func (n *Neuron) Address() string {
	return n.address
}

func (n *Neuron) Get(key string) interface{} {
	if n.values != nil {
		return n.values[key]
	}
	return nil
}

func (n *Neuron) Set(key string, value interface{}) {
	if n.values != nil {
		n.values = make(map[string]interface{})
	}
	n.values[key] = value
}

func (n *Neuron) Remove(key string) {
	if n.values != nil {
		delete(n.values, key)
	}
}
