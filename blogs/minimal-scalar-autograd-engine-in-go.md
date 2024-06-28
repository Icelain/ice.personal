---

date: 2024/06/28

---

# gograd - A minimal scalar autograd engine inspired by micrograd

I've recently seen a trend of implementing Andrej Karpathy's popular autograd engine [micrograd](https://github.com/karpathy/micrograd) in a variety of languages. I'd attribute this to the result of the increasing interest in understanding the fundamentals of neural networks, which is in conjunction to the ongoing AI hype train. 

Now there are probably already a couple of micrograd implementations in go, but as a relearning exercise, I've decided to create one directly from memory instead of replicating everything Karpathy did. This is based on what I remember was the significant logical portion of the original library. 

Why Go? Because it's simple, readable, and gets the job done. I could have probably used Rust for a challenge, but the objective here is to refresh on backpropagation as quickly and simply as possible, without any qualms with the borrow checker(Neuron is a recursive type so I would probably be Box<>ing everything to the heap).

Anyways, this article isn't going to be a long one, since this is extremely simple.

We have our Neuron struct:
```go
type Neuron struct {
	Value        float64
	Grad         float64
	Partner      *Neuron
	Constituents [2]*Neuron // index 0 is always (first) and index 1 is always (second)
	Operation    Op // Operation used to create this Neuron
}
```
It holds:
=> ```float64``` **Value**, which is the value of the Neuron,
=> ```float64``` **Grad**, which is the gradient of the Neuron,
=> A pointer to another neuron, which is its partner in in the mathematical operation; **Partner**
=> An array of pointers to the 2 Neurons which created the current Neuron; **Constituents**
=> The operation used to create the current Neuron; **Operation** 

The **Operation** field is just an enum of the type Op, 
defined as follows:
```go
type Op uint

const (
	NoneOp Op = iota
	AddOp
	MulOp
	PowOp
)
```
Now all we have to do is simply implement each of these operations:
```go
// Adds two neurons
func (n *Neuron) Add(to *Neuron) Neuron {

	n.Partner = to
	to.Partner = n

	res := Neuron{Value: n.Value + to.Value}
	res.Constituents = [2]*Neuron{n, to}

	res.Operation = AddOp

	return res
}

// Multiply's two neurons
func (n *Neuron) Mul(to *Neuron) Neuron {

	n.Partner = to
	to.Partner = n
	res := Neuron{Value: n.Value * to.Value}
	res.Constituents = [2]*Neuron{n, to}
	res.Operation = MulOp

	return res

}

// Pow() function for two neurons
//Here, the constituents have index 0 as the base and index 1 as the power
func (n *Neuron) Pow(to *Neuron) Neuron {

	n.Partner = to
	to.Partner = n
	res := Neuron{Value: math.Pow(n.Value, to.Value)}
	res.Constituents = [2]*Neuron{n, to}

	res.Operation = PowOp

	return res

}
```
Apart from the usual boilerplate, all that's left is a simple backpropagation method called Gradient():
```go
// Backprop function
func (n *Neuron) Gradient() {

	n.Grad = 1.0

	traverse(n.Constituents[0], n.Grad, n.Operation, true)
	traverse(n.Constituents[1], n.Grad, n.Operation, false)

}
```

The function ```traverse``` goes through every constituent neuron of each neuron and computes the gradients with respect to the different operations by using partial derivatives and the chain rule. This is done until there are no more constituents left for the neuron chain:

```go
// first^second
func traverse(n *Neuron, prevgrad float64, operation Op, first bool) {

	head := n
		
	if head==nil {

		return

	}

	switch operation {

	case AddOp:
		head.Grad = prevgrad

	case MulOp:
		head.Grad = prevgrad * (head.Partner.Value)

	case PowOp:
		if first {
			head.Grad = prevgrad * (math.Pow(head.Value, head.Partner.Value) * (head.Partner.Value / head.Value))
		} else {

			head.Grad = prevgrad * (math.Pow(head.Partner.Value, head.Value) * (math.Log(head.Partner.Value)))

		}

	}
```
**Note:** Notice how the order of the constituents matters when computing more complex functions like ```Pow()```.

And, that is all there is to it.

Before ending, I would like to point out that YES, there are design decisions taken here just for the sake of simplicity, such as the use of pointers for everything(which make reusing Neurons for operations not possible).

The repository(given below) will be updated sooner or later with more data structures like Nets, Layers and Multi-Layer Perceptrons.
 
Source code: [gograd](https://github.com/icelain/gograd)
