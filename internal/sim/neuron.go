package sim

type Neuron struct {
	Voltage, Threshold, Decay float64
}

func (n *Neuron) Step(input float64) bool {
	n.Voltage = n.Decay*n.Voltage + input
	if n.Voltage >= n.Threshold {
		n.Voltage = 0
		return true
	}
	return false
}
