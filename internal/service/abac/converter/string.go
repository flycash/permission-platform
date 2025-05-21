package converter

type StrConverter struct{}

func NewStrConverter() StrConverter {
	return StrConverter{}
}

func (s StrConverter) Decode(str string) (string, error) {
	return str, nil
}

func (s StrConverter) Encode(t string) (string, error) {
	return t, nil
}
