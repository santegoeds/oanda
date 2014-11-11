package analytics

import (
	"math"
	"strconv"
	"strings"
)

var (
	nan = math.NaN()
)

type Window struct {
	values []float64
}

// NewWindow returns a new window able to hold up to capacity values.
func NewWindow(capacity int) *Window {
	w := Window{
		values: make([]float64, 0, capacity),
	}
	return &w
}

// Len returns the number of values in the Window.
func (w Window) Len() int { return len(w.values) }

// Cap returns the maximum number of values that the Window can hold.
func (w Window) Cap() int { return cap(w.values) }

// Values returns the values as a slice of float64 values.
func (w Window) Values() []float64 { return w.values }

// String implements the fmt.Stringer interface.
func (w Window) String() string {
	join := func(fs []float64) string {
		ss := make([]string, len(fs))
		for _, f := range fs {
			ss = append(ss, strconv.FormatFloat(f, 'f', -1, 64))
		}
		return strings.Join(ss, ", ")
	}

	str := "Window{"
	if w.Len() < 10 {
		str += join(w.values)
	} else {
		str += join(w.values[:3]) + " ... " + join(w.values[len(w.values)-3:])
	}
	return str + "}"
}

// Push adds a new value to the front of the Window.  The Window's length increments until it
// the Window's capacity. If the Window is at full capacity existing values are shifted and the
// oldest values are discarded.
func (w *Window) Push(val ...float64) *Window {
	min := func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}

	// Number of values that need to be shifted towards the end
	// of the slice
	nShift := min(w.Cap()-len(val), w.Len())

	// Increase the size of the array.
	w.expandTo(w.Len() + len(val))

	// Shift values
	for i := 0; i < nShift; i++ {
		w.values[w.Len()-1-i] = w.values[nShift-1-i]
	}

	// skip
	if len(val) > w.Cap() {
		val = val[len(val)-w.Cap():]
	}

	idx := len(val) - 1
	for _, v := range val {
		w.values[idx] = v
		idx--
	}
	return w
}

// Sum returns the sum of all values in the Window.
func (w Window) Sum() float64 {
	sum := 0.0
	for _, f := range w.values {
		sum += f
	}
	return sum
}

// Slice returns a new Window that refers to a subrange of the original Window. Both Window's
// share the underlying data.
func (w Window) Slice(start, end int) *Window {
	if end > w.Len() {
		end = w.Len()
	}

	switch {
	case start < 0 && end < 0:
		return &w
	case start < 0:
		return &Window{w.values[:end]}
	case end < 0:
		return &Window{w.values[start:]}
	}

	return &Window{w.values[start:end]}
}

func (w Window) Clone() *Window {
	c := Window{
		values: make([]float64, w.Len(), w.Cap()),
	}
	for i := 0; i < w.Len(); i++ {
		c.values[i] = w.values[i]
	}
	return &c
}

func (w *Window) expandTo(n int) {
	if n > w.Cap() {
		n = w.Cap()
	}
	for i := w.Len(); i < n; i++ {
		w.values = append(w.values, nan)
	}
}
