package converter

type Converter[T any] interface {
	Decode(str string) (T, error)
	Encode(t T) (string, error)
}
