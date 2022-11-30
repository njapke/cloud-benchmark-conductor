package profile

import "fmt"

const (
	outTypeDotStr  = "dot"
	outTypeSVGStr  = "svg"
	outTypeTextStr = "text"
	outTypeGVStr   = "gv"
	outTypePSStr   = "ps"
	outTypeGIFStr  = "gif"
)

type OutType int

func OutTypeFrom(s string) (OutType, error) {
	var t OutType
	switch s {
	case outTypeDotStr:
		t = TypeDot
	case outTypeSVGStr:
		t = TypeSVG
	case outTypeTextStr:
		t = TypeText
	case outTypeGVStr:
		t = TypeGV
	case outTypePSStr:
		t = TypePS
	case outTypeGIFStr:
		t = TypeGIF
	default:
		return 0, fmt.Errorf("invalid OutType %s", s)
	}
	return t, nil
}

func (t OutType) Name() string {
	var s string
	switch t {
	case TypeDot:
		s = outTypeDotStr
	case TypeSVG:
		s = outTypeSVGStr
	case TypeText:
		s = outTypeTextStr
	case TypeGV:
		s = outTypeGVStr
	case TypePS:
		s = outTypePSStr
	case TypeGIF:
		s = outTypeGIFStr
	default:
		panic(fmt.Sprintf("invalid OutType %d", t))
	}
	return s
}

const (
	TypeDot = iota
	TypeSVG
	TypeText
	TypeGV
	TypePS
	TypeGIF
)

const (
	NodeCountDefault    = 80
	NodeFractionDefault = 0.005
	EdgeFractionDefault = 0.001
)

type OutConfig struct {
	Type         OutType // output file type
	NodeCount    int     // most important n nodes from profile
	NodeFraction float64 // drop least important x percent of nodes from profile
	EdgeFraction float64 // drop least important x percent of edges from profile
}

func (c OutConfig) FileSuffix() string {
	out := rdus(fmt.Sprintf("%d__%.5f__%.5f", c.NodeCount, c.NodeFraction, c.EdgeFraction))
	return fmt.Sprintf("%s.%s", out, c.Type.Name())
}
