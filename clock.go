package main

import (
	"fmt"
	"log"
	"math"
	"strings"
	"syscall/js"
)

const xmlns = `http://www.w3.org/2000/svg`

var _document js.Value
var _svg js.Value

func document() js.Value {
	if _document.IsUndefined() {
		_document = js.Global().Get(`document`)
	}
	return _document
}

func makeA(what string) SVGElement {
	return SVGElement(document().Call(`createElementNS`, xmlns, what))
}

func svg() js.Value {
	if _svg.IsUndefined() {
		svgs := document().Call(`getElementsByTagName`, `svg`)
		if svgs.IsUndefined() || svgs.Type() != js.TypeObject {
			log.Fatalf(`could not find any SVG elements`)
		}
		_svg = svgs.Index(0)
	}
	return _svg
}

type Stringer interface {
	String() string
}

type Coord float64

func (c Coord) String() string {
	return fmt.Sprintf(`%f`, c)
}

type BoolArg bool

func (b BoolArg) String() string {
	return map[BoolArg]string{
		false: `0`,
		true:  `1`,
	}[b]
}

type Point struct {
	X Coord
	Y Coord
}

func (p Point) String() string {
	return fmt.Sprintf(`%f,%f`, p.X, p.Y)
}

func PointFromPolar(r, theta Coord) Point {
	return Point{
		Coord(math.Cos(float64(theta))) * r,
		Coord(math.Sin(float64(theta))) * r,
	}
}

func (left Point) Add(right Point) Point {
	return Point{left.X + right.X, left.Y + right.Y}
}

type Instruction struct {
	Instruction rune
	Arguments   []Stringer
}

func NewInstruction(inst rune, arguments ...Stringer) Instruction {
	return Instruction{
		Instruction: inst,
		Arguments:   arguments,
	}
}

func (i Instruction) String() string {
	var builder strings.Builder
	builder.WriteRune(i.Instruction)
	for _, a := range i.Arguments {
		builder.WriteRune(' ')
		builder.WriteString(a.String())
	}
	return builder.String()
}

// A SVG Path object
type Path struct {
	instructions []Instruction
	lastPoint    Point
}

func NewPath() *Path {
	return &Path{
		instructions: make([]Instruction, 0),
	}
}

func (p *Path) updateLastPoint(pt Point) {
	p.lastPoint = pt
}

func (p *Path) moveLastPoint(pt Point) {
	p.lastPoint = p.lastPoint.Add(pt)
}

func (p *Path) Append(i Instruction) *Path {
	p.instructions = append(p.instructions, i)
	return p
}

// MoveTo lifts the pen and moves it to a new absolute location
func (p *Path) MoveTo(pt Point) *Path {
	p.updateLastPoint(pt)
	return p.Append(
		NewInstruction(
			'M',
			p.lastPoint,
		),
	)
}

// Move lifts the pen and moves it to a new relative location
func (p *Path) Move(ptd Point) *Path {
	p.moveLastPoint(ptd)
	return p.Append(
		NewInstruction(
			'm',
			p.lastPoint,
		),
	)
}

// MovePolar lifts the pen and moves it to a new relative location given by polar coordinates
func (p *Path) MovePolar(r, theta Coord) *Path {
	return p.Move(PointFromPolar(r, theta))
}

// LineTo lifts the pen and moves it to a new absolute location
func (p *Path) LineTo(pt Point) *Path {
	p.updateLastPoint(pt)
	return p.Append(
		NewInstruction(
			'L',
			p.lastPoint,
		),
	)
}

// Line lifts the pen and moves it to a new relative location
func (p *Path) Line(ptd Point) *Path {
	p.moveLastPoint(ptd)
	return p.Append(
		NewInstruction(
			'l',
			p.lastPoint,
		),
	)
}

// LinePolar lifts the pen and moves it to a new relative location given by polar coordinates
func (p *Path) LinePolar(r, theta Coord) *Path {
	return p.Line(PointFromPolar(r, theta))
}

// EndOpen ends the path, but doesn't connect it to the start.
func (p *Path) EndOpen() *Path {
	return p.Append(NewInstruction('z'))
}

// EndClosed ends the path, connecting it to the start.
func (p *Path) EndClosed() *Path {
	return p.Append(NewInstruction('Z'))
}

// Horizontal draws a horizontal line from the current point
func (p *Path) Horizontal(dx Coord) *Path {
	p.moveLastPoint(Point{dx, 0})
	return p.Append(NewInstruction('h', dx))
}

// HorizontalTo draws a horizontal line to the given X coordinate
func (p *Path) HorizontalTo(x Coord) *Path {
	p.updateLastPoint(Point{x, p.lastPoint.Y})
	return p.Append(NewInstruction('H', x))
}

// Vertical draws a vertical line from the current point
func (p *Path) Vertical(dy Coord) *Path {
	p.moveLastPoint(Point{0, dy})
	return p.Append(NewInstruction('v', dy))
}

// VerticalTo draws a horizontal line to the given X coordinate
func (p *Path) VerticalTo(y Coord) *Path {
	p.updateLastPoint(Point{p.lastPoint.X, y})
	return p.Append(NewInstruction('V', y))
}

// Cubic draws a cubic Bezier curve with relative coordinates
func (p *Path) Cubic(startControl, endControl, end Point) *Path {
	p.moveLastPoint(end)
	return p.Append(NewInstruction('c', startControl, endControl, end))
}

// CubicTo draws a cubic Bezier curve with absolute coordinates
func (p *Path) CubicTo(startControl, endControl, end Point) *Path {
	p.updateLastPoint(end)
	return p.Append(NewInstruction('C', startControl, endControl, end))
}

// ContinueCubic continues a smooth line of cubic Bezier curves with relative coordinates
func (p *Path) ContinueCubic(endControl, end Point) *Path {
	p.moveLastPoint(end)
	return p.Append(NewInstruction('s', endControl, end))
}

// ContinueCubicTo continues a smooth line of cubic Bezier curves with absolute coordinates
func (p *Path) ContinueCubicTo(endControl, end Point) *Path {
	p.updateLastPoint(end)
	return p.Append(NewInstruction('S', endControl, end))
}

// Quadratic draws a quadratic curve with relative coordinates
func (p *Path) Quadratic(control, end Point) *Path {
	p.moveLastPoint(end)
	return p.Append(NewInstruction('q', control, end))
}

// QuadraticTo draws a quadratic curve with absolute coordinates
func (p *Path) QuadraticTo(control, end Point) *Path {
	p.updateLastPoint(end)
	return p.Append(NewInstruction('Q', control, end))
}

// ContinueQuadratic continues a smooth line of quadratics with relative coordinates
func (p *Path) ContinueQuadratic(end Point) *Path {
	p.moveLastPoint(end)
	return p.Append(NewInstruction('t', end))
}

// ContinueQuadraticTo continues a smooth line of quadratics with absolute coordinates
func (p *Path) ContinueQuadraticTo(end Point) *Path {
	p.updateLastPoint(end)
	return p.Append(NewInstruction('T', end))
}

var boolToCoord = map[bool]Coord{false: 0, true: 1}

// Arc draws an arc with relative coordinates.
// rx and ry are the radii of the ellipse
// rot is the rotation of the ellipse (in degrees)
// largeArc is whether to take the shortest path (false) or the longest (true)
// sweep is whether to begin with negative angles (false) or positive (true)
// end is where to end the arc, in relative coordinates
func (p *Path) Arc(rx, ry, rot Coord, largeArc, sweep bool, end Point) *Path {
	p.moveLastPoint(end)

	return p.Append(NewInstruction(
		'a',
		rx, ry,
		rot,
		boolToCoord[largeArc],
		boolToCoord[sweep],
		end,
	))
}

// ArcTo draws an arc with absolute coordinates.
// rx and ry are the radii of the ellipse
// rot is the rotation of the ellipse (in degrees)
// largeArc is whether to take the shortest path (false) or the longest (true)
// sweep is whether to begin with negative angles (false) or positive (true)
// end is where to end the arc, in absolute coordinates
func (p *Path) ArcTo(rx, ry, rot Coord, largeArc, sweep bool, end Point) *Path {
	p.updateLastPoint(end)

	return p.Append(NewInstruction(
		'A',
		rx, ry,
		rot,
		BoolArg(largeArc),
		BoolArg(sweep),
		end,
	))
}

func (p *Path) String() string {
	var builder strings.Builder
	for _, inst := range p.instructions {
		builder.WriteString(inst.String())
	}
	return builder.String()
}

func (p *Path) toElement() SVGElement {
	el := makeA(`path`)
	js.Value(el).Call(`setAttribute`, `d`, p.String())
	return el
}

type SVGElement js.Value
type Attributes map[string]interface{}

func (s SVGElement) SetAttribute(name string, value interface{}) {
	js.Value(s).Call(`setAttribute`, name, value)
}

func (s SVGElement) SetAttributes(atts Attributes) {
	for key, val := range atts {
		s.SetAttribute(key, val)
	}
}

func (s SVGElement) AppendChild(child js.Value) {
	js.Value(s).Call(`appendChild`, child)
}

func Write(text string, x, y Coord, size string) SVGElement {
	tnode := document().Call(`createTextNode`, text)
	tel := makeA(`text`)
	tel.AppendChild(tnode)
	tel.SetAttributes(Attributes{
		`font-size`:          size,
		`x`:                  x,
		`y`:                  y,
		`text-anchor`:        `middle`,
		`alignment-baseline`: `middle`,
	})
	return tel
}

func getAttributes(atts js.Value) Attributes {
	aa := make(Attributes)
	object := js.Global().Get(`Object`)
	entries := object.Call(`entries`, atts)
	numEntries := entries.Length()
	for i := 0; i < numEntries; i++ {
		entry := entries.Index(i)
		key := entry.Index(0).String()
		val := entry.Index(1).String()
		aa[key] = val
	}
	return aa
}

/* Chrome removed the ability to style elements referred to
 * by a <use> element, because they have a better plan to replace
 * it ... which they haven't implemented yet. (ノಠ益ಠ)ノ彡┻━┻
 */
func realisesUses(_ js.Value, _ []js.Value) interface{} {
	uses := document().Call(`getElementsByTagName`, `use`)

	if uses.IsUndefined() || uses.Type() != js.TypeObject {
		log.Fatalf(`could not find any use elements`)
	}
	numUses := uses.Length()
	for i := 0; i < numUses; i++ {
		use := uses.Index(i)
		href := use.Call(`getAttribute`, `href`).String()
		id := strings.Split(href, `#`)[1]
		el := document().Call(`getElementBId`, id)
		clone := el.Call(`cloneNode`, true)
		attributes := getAttributes(use.Get(`attributes`))
		delete(attributes, `href`)
		clone.Call(`removeAttribute`, `id`)
		SVGElement(clone).SetAttributes(attributes)
		use.Get(`parentElement`).Call(`insertBefore`, clone, use)
		use.Call(`remove`)
	}
	return nil
}

func Init() {
	js.Global().Set(`realizeUses`, js.FuncOf(realisesUses))
}
