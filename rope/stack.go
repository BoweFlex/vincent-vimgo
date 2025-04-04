package rope

type Stack []interface{}

func (s *Stack) Push(i interface{}) {
	*s = append(*s, i)
}

func (s *Stack) IsEmpty() bool {
	return len(*s) == 0
}

func (s *Stack) Size() int {
	return len(*s)
}

func (s *Stack) Pop() (last interface{}) {
	last = s.Peek()
	if last != nil {
		// Remove last from the stack
		*s = (*s)[:len(*s)-1]
	}
	return
}

func (s *Stack) Peek() (last interface{}) {
	if s.IsEmpty() {
		return nil
	}
	last = (*s)[len(*s)-1]
	return
}
