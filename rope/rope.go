package rope

var (
	tokenLength int = 2
)

type Rope struct {
	Left     *Rope
	Right    *Rope
	Parent   *Rope
	Contents string
	Length   int
}

func NewRope(node, parent *Rope, input string) Rope {
	temp_rope := &Rope{
		Left:     nil,
		Right:    nil,
		Parent:   parent,
		Contents: "",
		Length:   0,
	}

	node = temp_rope
	if len(input) > tokenLength {
		node.Length = int(len(input) / tokenLength)
		left := NewRope(node.Left, node, input[0:node.Length])
		right := NewRope(node.Right, node, input[node.Length:])
		node.Left = &left
		node.Right = &right
	} else {
		node.Length = len(input)
		for _, char := range input {
			node.Contents += string(char)
		}
	}

	return *node
}

func (r *Rope) GetString() string {
	if r.Left == nil && r.Right == nil {
		return r.Contents
	}
	return (r.Left.GetString() + r.Right.GetString())
}

func ConcatRopes(left, right *Rope) (r Rope) {
	r.Parent, r.Contents = nil, ""
	r.Left, r.Right = left, right
	r.Length = left.Length
	left.Parent, right.Parent = &r, &r
	return
}
