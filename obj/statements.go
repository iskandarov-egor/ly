package obj

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"strconv"
	"unsafe"
)

type ObjFile struct {
	Statements []Statement
	VStatements []*VStatement
	VtStatements []*VtStatement
	VnStatements []*VnStatement
}

const (
	KeywordV  = 1010 + iota
	KeywordVt = 1010 + iota
	KeywordVn = 1010 + iota
	KeywordO  = 1010 + iota
	KeywordF  = 1010 + iota
	KeywordUsemtl = 1010 + iota
	KeywordNewmtl = 1010 + iota
	KeywordMapKd  = 1010 + iota
)

type Statement interface {
	//Keyword() int
}

type VStatement struct {
	X, Y, Z float32
}

type VnStatement struct {
	VStatement
}

type OStatement struct {
	Name string
}

type UsemtlStatement struct {
	Name string
}

type NewmtlStatement struct {
	Name string
}

type MapKdStatement struct {
	Path string
}

type VtStatement struct {
	U, V float32
}

type FStatementPoint struct {
	V, Vt, Vn int
}

type FStatement struct {
	Points []FStatementPoint
}

func (s *VStatement) Keyword() int { return KeywordV; }

func (s *OStatement) Keyword() int { return KeywordO; }

func (s *VtStatement) Keyword() int { return KeywordVt; }

func (s *VnStatement) Keyword() int { return KeywordVn; }

func (s *FStatement) Keyword() int { return KeywordF; }

func str2float(s string) (float32) {
	ret, err := strconv.ParseFloat(s, 32)
	if err != nil {
		panic("aaa")
	}
	return float32(ret)
}

func checkFloats(s []string) (error) {
	for _, s := range s {
		_, err := strconv.ParseFloat(s, 32)
		if err != nil {
			return err
		}
	}
	return nil
}

func ParseV(words []string) (*VStatement, error) {
	if len(words) != 3 {
		fmt.Println(1, len(words))
		return nil, fmt.Errorf("expected 3 words in v statement")
	}
	if err := checkFloats(words); err != nil {
		return nil, fmt.Errorf("convert string to float: %s", err)
	}
	ret := VStatement{
		str2float(words[0]),
		str2float(words[1]),
		str2float(words[2]),
	}
	return &ret, nil
}

func ParseVn(words []string) (*VnStatement, error) {
	vt, err := ParseV(words)
	return (*VnStatement)(unsafe.Pointer(vt)), err
}

func ParseO(words []string) (*OStatement) {
	if len(words) != 1 {
		return nil
	}
	return &OStatement{words[0]}
}

func ParseUsemtl(words []string) (*UsemtlStatement) {
	return (*UsemtlStatement)(unsafe.Pointer(ParseO(words)))
}

func ParseNewmtl(words []string) (*NewmtlStatement) {
	return (*NewmtlStatement)(unsafe.Pointer(ParseO(words)))
}

func ParseMapKd(words []string) (*MapKdStatement) {
	return (*MapKdStatement)(unsafe.Pointer(ParseO(words)))
}

func ParseVt(words []string) (*VtStatement, error) {
	if len(words) < 2 {
		return nil, fmt.Errorf("expected 2 words in vt statement, got %d", len(words))
	}
	if nil != checkFloats(words) {
		return nil, fmt.Errorf("bad float in vt statement: %v", words)
	}
	ret := VtStatement{
		str2float(words[0]),
		str2float(words[1]),
	}
	return &ret, nil
}

func ParseF(words []string) (*FStatement, error) {
	if len(words) == 0 {
		return nil, nil
	}
	ret := FStatement{make([]FStatementPoint, len(words))}
	for i, word := range words {
		elems := strings.Split(word, "/")
		var err error
		ret.Points[i].V, err = strconv.Atoi(elems[0])
		if err != nil {
			return nil, fmt.Errorf("parse face point %d: %v", i + 1, err)
		}
		if len(elems) == 3 {
			if elems[1] != "" {
				ret.Points[i].Vt, err = strconv.Atoi(elems[1])
				if err != nil {
					return nil, fmt.Errorf("parse face point %d: %v", i + 1, err)
				}
			}
			if elems[2] != "" {
				ret.Points[i].Vn, err = strconv.Atoi(elems[2])
				if err != nil {
					return nil, fmt.Errorf("parse face point %d: %v", i + 1, err)
				}
			}
		}
		if len(elems) == 2 && elems[1] != "" {
			ret.Points[i].Vn, err = strconv.Atoi(elems[1])
			if err != nil {
				return nil, fmt.Errorf("parse face point %d: %v", i + 1, err)
			}
		}
	}
	return &ret, nil
}

func ParseObj(path string) *ObjFile {
	obj := ObjFile{}
	file, _ := os.Open(path)
	scanner := bufio.NewScanner(file)
	var statements []Statement
	lineno := 1
	for scanner.Scan() {
		txt := strings.TrimSpace(scanner.Text())
		for strings.HasSuffix(txt, "\\") && scanner.Scan() {
			txt = strings.TrimRight(txt, "\\")
			// line wrapping
			txt += strings.TrimSpace(scanner.Text())
		}
		words := strings.Fields(txt)
		if len(words) == 0 {
			continue
		}
		
		var statement Statement
		var parseErr error
		switch words[0] {
			case "v":
				v, err := ParseV(words[1:])
				parseErr = err
				if v != nil {
					obj.VStatements = append(obj.VStatements, v)
					statement = v
				}
			case "o":
				statement = ParseO(words[1:])
			case "vt":
				var vt *VtStatement
				vt, parseErr = ParseVt(words[1:])
				if vt != nil {
					obj.VtStatements = append(obj.VtStatements, vt)
					statement = vt
				}
			case "vn":
				vn, err := ParseVn(words[1:])
				parseErr = err
				if vn != nil {
					obj.VnStatements = append(obj.VnStatements, vn)
					statement = vn
				}
			case "f":
				var f *FStatement
				f, parseErr = ParseF(words[1:])
				if f != nil {
					statement = f
				}
			case "usemtl":
				statement = ParseUsemtl(words[1:])
			case "newmtl":
				statement = ParseNewmtl(words[1:])
			case "map_Kd":
				statement = ParseMapKd(words[1:])
			default:
				lineno++
				continue
		}
		if parseErr != nil {
			panic(fmt.Errorf("parse error on line %d: %v", lineno, parseErr))
		}
		statements = append(statements, statement)
		lineno++
	}
	obj.Statements = statements
	return &obj
}

func main() {
}
