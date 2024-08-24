// This package implements Maybe type.
//
// It is inspired by Rust's Optional<T> and https://github.com/pmorelli92/maybe, though this
// version contains a bit more methods compared to the latter.

package maybe

import (
	"encoding/json"
)

type Maybe[T any] struct {
	some bool
	v    T
}

func (m Maybe[T]) IsSome() bool {
	return m.some
}

func (m Maybe[T]) IsNone() bool {
	return !m.some
}

func (m Maybe[T]) Get() T {
	return m.v
}

func (m Maybe[T]) TryGet() (T, bool) {
	return m.v, m.some
}

func (m Maybe[T]) GetOr(def T) T {
	if !m.some {
		return def
	}
	return m.v
}

func Some[T any](v T) Maybe[T] {
	return Maybe[T]{some: true, v: v}
}

func None[T any]() Maybe[T] {
	return Maybe[T]{some: false}
}

func (m *Maybe[T]) UnmarshalJSON(data []byte) error {
	var v *T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if v == nil {
		*m = None[T]()
	} else {
		*m = Some(*v)
	}
	return nil
}

func (m Maybe[T]) MarshalJSON() ([]byte, error) {
	var v *T
	if m.some {
		v = &m.v
	}
	return json.Marshal(v)
}
