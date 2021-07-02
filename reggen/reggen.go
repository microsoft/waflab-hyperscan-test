// Package reggen generates text based on regex definitions
package reggen

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"
)

const runeRangeEnd = 0x10ffff
const printableChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

var printableCharsNoNL = printableChars[:len(printableChars)-2]

type state struct {
	limit int
}

type Generator struct {
	re    *Regexp
	rand  *rand.Rand
	debug bool
}

func max(num1, num2 int) int {
	if num1 > num2 {
		return num1
	}
	return num2
}

func Length(re *Regexp) int {
	op := re.Op
	switch op {
	case OpNoMatch:
	case OpEmptyMatch:
	case OpLiteral:
		return len(re.Rune)
	case OpCharClass:
		return 1
	case OpAnyCharNotNL, OpAnyChar:
		return 1
	case OpBeginLine:
	case OpEndLine:
	case OpBeginText:
	case OpEndText:
	case OpWordBoundary:
	case OpNoWordBoundary:
	case OpCapture:
		return Length(re.Sub0[0])
	case OpStar, OpPlus, OpQuest, OpRepeat, OpConcat:
		l := 0
		for _, r := range re.Sub {
			l += Length(r)
		}
		return l * 10
	case OpAlternate:
		l := 0
		for _, r := range re.Sub {
			l = max(l, Length(r))
		}
		return l
	default:
		panic("unhandled operator")
	}
	return 0
}

func (g *Generator) generate(s *state, re *Regexp) string {
	op := re.Op
	switch op {
	case OpNoMatch:
	case OpEmptyMatch:
		return ""
	case OpLiteral:
		res := ""
		for _, r := range re.Rune {
			res += string(r)
		}
		return res
	case OpCharClass:
		// number of possible chars
		sum := 0
		for i := 0; i < len(re.Rune); i += 2 {
			if g.debug {
				fmt.Printf("Range: %#U-%#U\n", re.Rune[i], re.Rune[i+1])
			}
			sum += int(re.Rune[i+1]-re.Rune[i]) + 1
			if re.Rune[i+1] == runeRangeEnd {
				sum = -1
				break
			}
		}
		// pick random char in range (inverse match group)
		if sum == -1 {
			possibleChars := []uint8{}
			for j := 0; j < len(printableChars); j++ {
				c := printableChars[j]
				//fmt.Printf("Char %c %d\n", c, c)
				// Check c in range
				for i := 0; i < len(re.Rune); i += 2 {
					if rune(c) >= re.Rune[i] && rune(c) <= re.Rune[i+1] {
						possibleChars = append(possibleChars, c)
						break
					}
				}
			}
			//fmt.Println("Possible chars: ", possibleChars)
			if len(possibleChars) > 0 {
				c := possibleChars[g.rand.Intn(len(possibleChars))]
				if g.debug {
					fmt.Printf("Generated rune %c for inverse range %v\n", c, re)
				}
				return string([]byte{c})
			}
		}
		if g.debug {
			fmt.Println("Char range: ", sum)
		}
		r := g.rand.Intn(int(sum))
		var ru rune
		sum = 0
		for i := 0; i < len(re.Rune); i += 2 {
			gap := int(re.Rune[i+1]-re.Rune[i]) + 1
			if sum+gap > r {
				ru = re.Rune[i] + rune(r-sum)
				break
			}
			sum += gap
		}
		if g.debug {
			fmt.Printf("Generated rune %c for range %v\n", ru, re)
		}
		return string(ru)
	case OpAnyCharNotNL, OpAnyChar:
		chars := printableChars
		if op == OpAnyCharNotNL {
			chars = printableCharsNoNL
		}
		c := chars[g.rand.Intn(len(chars))]
		return string([]byte{c})
	case OpBeginLine:
	case OpEndLine:
	case OpBeginText:
	case OpEndText:
	case OpWordBoundary:
	case OpNoWordBoundary:
	case OpCapture:
		if g.debug {
			fmt.Println("OpCapture", re.Sub, len(re.Sub))
		}
		return g.generate(s, re.Sub0[0])
	case OpStar:
		// Repeat zero or more times
		res := ""
		count := g.rand.Intn(s.limit + 1)
		for i := 0; i < count; i++ {
			for _, r := range re.Sub {
				res += g.generate(s, r)
			}
		}
		return res
	case OpPlus:
		// Repeat one or more times
		res := ""
		count := g.rand.Intn(s.limit) + 1
		for i := 0; i < count; i++ {
			for _, r := range re.Sub {
				res += g.generate(s, r)
			}
		}
		return res
	case OpQuest:
		// Zero or one instances
		res := ""
		count := g.rand.Intn(2)
		if g.debug {
			fmt.Println("Quest", count)
		}
		for i := 0; i < count; i++ {
			for _, r := range re.Sub {
				res += g.generate(s, r)
			}
		}
		return res
	case OpRepeat:
		// Repeat one or more times
		if g.debug {
			fmt.Println("OpRepeat", re.Min, re.Max)
		}
		res := ""
		count := 0
		re.Max = int(math.Min(float64(re.Max), float64(s.limit)))
		if re.Max > re.Min {
			count = g.rand.Intn(re.Max - re.Min + 1)
		}
		if g.debug {
			fmt.Println(re.Max, count)
		}
		for i := 0; i < re.Min || i < (re.Min+count); i++ {
			for _, r := range re.Sub {
				res += g.generate(s, r)
			}
		}
		return res
	case OpConcat:
		// Concatenate sub-regexes
		res := ""
		for _, r := range re.Sub {
			res += g.generate(s, r)
		}
		return res
	case OpAlternate:
		if g.debug {
			fmt.Println("OpAlternative", re.Sub, len(re.Sub))
		}
		index, value := 0, 0
		for i, r := range re.Sub {
			curr := Length(r)
			if curr > value {
				index = i
				value = curr
			}
		}
		return g.generate(s, re.Sub[index])
	default:
		fmt.Fprintln(os.Stderr, "[reg-gen] Unhandled op: ", op)
	}
	return ""
}

// limit is the maximum number of times star, range or plus should repeat
// i.e. [0-9]+ will generate at most 10 characters if this is set to 10
func (g *Generator) Generate(limit int) string {
	return g.generate(&state{limit: limit}, g.re)
}

// create a new generator
func NewGenerator(regex string) (*Generator, error) {
	re, err := Parse(regex, Perl)
	if err != nil {
		return nil, err
	}
	//fmt.Println("Compiled re ", re)
	return &Generator{
		re:   re,
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

func (gen *Generator) SetSeed(seed int64) {
	gen.rand = rand.New(rand.NewSource(seed))
}

func Generate(regex string, limit int) (string, error) {
	g, err := NewGenerator(regex)
	if err != nil {
		return "", err
	}
	return g.Generate(limit), nil
}
