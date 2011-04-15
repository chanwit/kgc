// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The printer package implements printing of AST nodes.
package printer

import (
	"bytes"
	"fmt"
	"korat/ast"
	"korat/token"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"tabwriter"
)


const debug = false // enable for debugging


type whiteSpace int

const (
	ignore   = whiteSpace(0)
	blank    = whiteSpace(' ')
	vtab     = whiteSpace('\v')
	newline  = whiteSpace('\n')
	formfeed = whiteSpace('\f')
	indent   = whiteSpace('>')
	unindent = whiteSpace('<')
)


const (
	esc2 = '\xfe'                        // an escape byte that cannot occur in regular UTF-8
	_    = 1 / (esc2 - tabwriter.Escape) // cause compiler error if esc2 == tabwriter.Escape
)


var (
	esc       = []byte{tabwriter.Escape}
	htab      = []byte{'\t'}
	htabs     = []byte("\t\t\t\t\t\t\t\t")
	newlines  = []byte("\n\n\n\n\n\n\n\n") // more than the max determined by nlines
	formfeeds = []byte("\f\f\f\f\f\f\f\f") // more than the max determined by nlines
)


// Special positions
var noPos token.Position // use noPos when a position is needed but not known
var infinity = 1 << 30


// Use ignoreMultiLine if the multiLine information is not important.
var ignoreMultiLine = new(bool)


// A pmode value represents the current printer mode.
type pmode int

const (
	inLiteral pmode = 1 << iota
	noExtraLinebreak
)


type printer struct {
	// Configuration (does not change after initialization)
	output io.Writer
	Config
	fset   *token.FileSet
	errors chan os.Error

	// Current state
	nesting int         // nesting level (0: top-level (package scope), >0: functions/decls.)
	written int         // number of bytes written
	indent  int         // current indentation
	mode    pmode       // current printer mode
	lastTok token.Token // the last token printed (token.ILLEGAL if it's whitespace)

	// Buffered whitespace
	buffer []whiteSpace

	// The (possibly estimated) position in the generated output;
	// in AST space (i.e., pos is set whenever a token position is
	// known accurately, and updated dependending on what has been
	// written).
	pos token.Position

	// The value of pos immediately after the last item has been
	// written using writeItem.
	last token.Position

	// HTML support
	lastTaggedLine int // last line for which a line tag was written

	// The list of all source comments, in order of appearance.
	comments        []*ast.CommentGroup // may be nil
	cindex          int                 // current comment index
	useNodeComments bool                // if not set, ignore lead and line comments of nodes
}


func (p *printer) init(output io.Writer, cfg *Config, fset *token.FileSet) {
	p.output = output
	p.Config = *cfg
	p.fset = fset
	p.errors = make(chan os.Error)
	p.buffer = make([]whiteSpace, 0, 16) // whitespace sequences are short
}


func (p *printer) internalError(msg ...interface{}) {
	if debug {
		fmt.Print(p.pos.String() + ": ")
		fmt.Println(msg...)
		panic("go/printer")
	}
}


// nlines returns the adjusted number of linebreaks given the desired number
// of breaks n such that min <= result <= max where max depends on the current
// nesting level.
//
func (p *printer) nlines(n, min int) int {
	if n < min {
		return min
	}
	max := 3 // max. number of newlines at the top level (p.nesting == 0)
	if p.nesting > 0 {
		max = 2 // max. number of newlines everywhere else
	}
	if n > max {
		return max
	}
	return n
}


// write0 writes raw (uninterpreted) data to p.output and handles errors.
// write0 does not indent after newlines, and does not HTML-escape or update p.pos.
//
func (p *printer) write0(data []byte) {
	if len(data) > 0 {
		n, err := p.output.Write(data)
		p.written += n
		if err != nil {
			p.errors <- err
			runtime.Goexit()
		}
	}
}


// write interprets data and writes it to p.output. It inserts indentation
// after a line break unless in a tabwriter escape sequence.
// It updates p.pos as a side-effect.
//
func (p *printer) write(data []byte) {
	i0 := 0
	for i, b := range data {
		switch b {
		case '\n', '\f':
			// write segment ending in b
			p.write0(data[i0 : i+1])

			// update p.pos
			p.pos.Offset += i + 1 - i0
			p.pos.Line++
			p.pos.Column = 1

			if p.mode&inLiteral == 0 {
				// write indentation
				// use "hard" htabs - indentation columns
				// must not be discarded by the tabwriter
				j := p.indent
				for ; j > len(htabs); j -= len(htabs) {
					p.write0(htabs)
				}
				p.write0(htabs[0:j])

				// update p.pos
				p.pos.Offset += p.indent
				p.pos.Column += p.indent
			}

			// next segment start
			i0 = i + 1

		case tabwriter.Escape:
			p.mode ^= inLiteral

			// ignore escape chars introduced by printer - they are
			// invisible and must not affect p.pos (was issue #1089)
			p.pos.Offset--
			p.pos.Column--
		}
	}

	// write remaining segment
	p.write0(data[i0:])

	// update p.pos
	d := len(data) - i0
	p.pos.Offset += d
	p.pos.Column += d
}


func (p *printer) writeNewlines(n int, useFF bool) {
	if n > 0 {
		n = p.nlines(n, 0)
		if useFF {
			p.write(formfeeds[0:n])
		} else {
			p.write(newlines[0:n])
		}
	}
}


// writeItem writes data at position pos. data is the text corresponding to
// a single lexical token, but may also be comment text. pos is the actual
// (or at least very accurately estimated) position of the data in the original
// source text. writeItem updates p.last to the position immediately following
// the data.
//
func (p *printer) writeItem(pos token.Position, data []byte) {
	if pos.IsValid() {
		// continue with previous position if we don't have a valid pos
		if p.last.IsValid() && p.last.Filename != pos.Filename {
			// the file has changed - reset state
			// (used when printing merged ASTs of different files
			// e.g., the result of ast.MergePackageFiles)
			p.indent = 0
			p.mode = 0
			p.buffer = p.buffer[0:0]
		}
		p.pos = pos
	}
	if debug {
		// do not update p.pos - use write0
		_, filename := filepath.Split(pos.Filename)
		p.write0([]byte(fmt.Sprintf("[%s:%d:%d]", filename, pos.Line, pos.Column)))
	}
	p.write(data)
	p.last = p.pos
}


// writeCommentPrefix writes the whitespace before a comment.
// If there is any pending whitespace, it consumes as much of
// it as is likely to help position the comment nicely.
// pos is the comment position, next the position of the item
// after all pending comments, prev is the previous comment in
// a group of comments (or nil), and isKeyword indicates if the
// next item is a keyword.
//
func (p *printer) writeCommentPrefix(pos, next token.Position, prev *ast.Comment, isKeyword bool) {
	if p.written == 0 {
		// the comment is the first item to be printed - don't write any whitespace
		return
	}

	if pos.IsValid() && pos.Filename != p.last.Filename {
		// comment in a different file - separate with newlines (writeNewlines will limit the number)
		p.writeNewlines(10, true)
		return
	}

	if pos.Line == p.last.Line && (prev == nil || prev.Text[1] != '/') {
		// comment on the same line as last item:
		// separate with at least one separator
		hasSep := false
		if prev == nil {
			// first comment of a comment group
			j := 0
			for i, ch := range p.buffer {
				switch ch {
				case blank:
					// ignore any blanks before a comment
					p.buffer[i] = ignore
					continue
				case vtab:
					// respect existing tabs - important
					// for proper formatting of commented structs
					hasSep = true
					continue
				case indent:
					// apply pending indentation
					continue
				}
				j = i
				break
			}
			p.writeWhitespace(j)
		}
		// make sure there is at least one separator
		if !hasSep {
			if pos.Line == next.Line {
				// next item is on the same line as the comment
				// (which must be a /*-style comment): separate
				// with a blank instead of a tab
				p.write([]byte{' '})
			} else {
				p.write(htab)
			}
		}

	} else {
		// comment on a different line:
		// separate with at least one line break
		if prev == nil {
			// first comment of a comment group
			j := 0
			for i, ch := range p.buffer {
				switch ch {
				case blank, vtab:
					// ignore any horizontal whitespace before line breaks
					p.buffer[i] = ignore
					continue
				case indent:
					// apply pending indentation
					continue
				case unindent:
					// if the next token is a keyword, apply the outdent
					// if it appears that the comment is aligned with the
					// keyword; otherwise assume the outdent is part of a
					// closing block and stop (this scenario appears with
					// comments before a case label where the comments
					// apply to the next case instead of the current one)
					if isKeyword && pos.Column == next.Column {
						continue
					}
				case newline, formfeed:
					// TODO(gri): may want to keep formfeed info in some cases
					p.buffer[i] = ignore
				}
				j = i
				break
			}
			p.writeWhitespace(j)
		}
		// use formfeeds to break columns before a comment;
		// this is analogous to using formfeeds to separate
		// individual lines of /*-style comments - but make
		// sure there is at least one line break if the previous
		// comment was a line comment
		n := pos.Line - p.last.Line // if !pos.IsValid(), pos.Line == 0, and n will be 0
		if n <= 0 && prev != nil && prev.Text[1] == '/' {
			n = 1
		}
		p.writeNewlines(n, true)
	}
}


func (p *printer) writeCommentLine(comment *ast.Comment, pos token.Position, line []byte) {
	// line must pass through unchanged, bracket it with tabwriter.Escape
	line = bytes.Join([][]byte{esc, line, esc}, nil)
	p.writeItem(pos, line)
}


// Split comment text into lines
func split(text []byte) [][]byte {
	// count lines (comment text never ends in a newline)
	n := 1
	for _, c := range text {
		if c == '\n' {
			n++
		}
	}

	// split
	lines := make([][]byte, n)
	n = 0
	i := 0
	for j, c := range text {
		if c == '\n' {
			lines[n] = text[i:j] // exclude newline
			i = j + 1            // discard newline
			n++
		}
	}
	lines[n] = text[i:]

	return lines
}


func isBlank(s []byte) bool {
	for _, b := range s {
		if b > ' ' {
			return false
		}
	}
	return true
}


func commonPrefix(a, b []byte) []byte {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] && (a[i] <= ' ' || a[i] == '*') {
		i++
	}
	return a[0:i]
}


func stripCommonPrefix(lines [][]byte) {
	if len(lines) < 2 {
		return // at most one line - nothing to do
	}
	// len(lines) >= 2

	// The heuristic in this function tries to handle a few
	// common patterns of /*-style comments: Comments where
	// the opening /* and closing */ are aligned and the
	// rest of the comment text is aligned and indented with
	// blanks or tabs, cases with a vertical "line of stars"
	// on the left, and cases where the closing */ is on the
	// same line as the last comment text.

	// Compute maximum common white prefix of all but the first,
	// last, and blank lines, and replace blank lines with empty
	// lines (the first line starts with /* and has no prefix).
	// In case of two-line comments, consider the last line for
	// the prefix computation since otherwise the prefix would
	// be empty.
	//
	// Note that the first and last line are never empty (they
	// contain the opening /* and closing */ respectively) and
	// thus they can be ignored by the blank line check.
	var prefix []byte
	if len(lines) > 2 {
		for i, line := range lines[1 : len(lines)-1] {
			switch {
			case isBlank(line):
				lines[1+i] = nil // range starts at line 1
			case prefix == nil:
				prefix = commonPrefix(line, line)
			default:
				prefix = commonPrefix(prefix, line)
			}
		}
	} else { // len(lines) == 2
		line := lines[1]
		prefix = commonPrefix(line, line)
	}

	/*
	 * Check for vertical "line of stars" and correct prefix accordingly.
	 */
	lineOfStars := false
	if i := bytes.Index(prefix, []byte{'*'}); i >= 0 {
		// Line of stars present.
		if i > 0 && prefix[i-1] == ' ' {
			i-- // remove trailing blank from prefix so stars remain aligned
		}
		prefix = prefix[0:i]
		lineOfStars = true
	} else {
		// No line of stars present.
		// Determine the white space on the first line after the /*
		// and before the beginning of the comment text, assume two
		// blanks instead of the /* unless the first character after
		// the /* is a tab. If the first comment line is empty but
		// for the opening /*, assume up to 3 blanks or a tab. This
		// whitespace may be found as suffix in the common prefix.
		first := lines[0]
		if isBlank(first[2:]) {
			// no comment text on the first line:
			// reduce prefix by up to 3 blanks or a tab
			// if present - this keeps comment text indented
			// relative to the /* and */'s if it was indented
			// in the first place
			i := len(prefix)
			for n := 0; n < 3 && i > 0 && prefix[i-1] == ' '; n++ {
				i--
			}
			if i == len(prefix) && i > 0 && prefix[i-1] == '\t' {
				i--
			}
			prefix = prefix[0:i]
		} else {
			// comment text on the first line
			suffix := make([]byte, len(first))
			n := 2 // start after opening /*
			for n < len(first) && first[n] <= ' ' {
				suffix[n] = first[n]
				n++
			}
			if n > 2 && suffix[2] == '\t' {
				// assume the '\t' compensates for the /*
				suffix = suffix[2:n]
			} else {
				// otherwise assume two blanks
				suffix[0], suffix[1] = ' ', ' '
				suffix = suffix[0:n]
			}
			// Shorten the computed common prefix by the length of
			// suffix, if it is found as suffix of the prefix.
			if bytes.HasSuffix(prefix, suffix) {
				prefix = prefix[0 : len(prefix)-len(suffix)]
			}
		}
	}

	// Handle last line: If it only contains a closing */, align it
	// with the opening /*, otherwise align the text with the other
	// lines.
	last := lines[len(lines)-1]
	closing := []byte("*/")
	i := bytes.Index(last, closing)
	if isBlank(last[0:i]) {
		// last line only contains closing */
		var sep []byte
		if lineOfStars {
			// insert an aligning blank
			sep = []byte{' '}
		}
		lines[len(lines)-1] = bytes.Join([][]byte{prefix, closing}, sep)
	} else {
		// last line contains more comment text - assume
		// it is aligned like the other lines
		prefix = commonPrefix(prefix, last)
	}

	// Remove the common prefix from all but the first and empty lines.
	for i, line := range lines[1:] {
		if len(line) != 0 {
			lines[1+i] = line[len(prefix):] // range starts at line 1
		}
	}
}


func (p *printer) writeComment(comment *ast.Comment) {
	text := comment.Text

	// shortcut common case of //-style comments
	if text[1] == '/' {
		p.writeCommentLine(comment, p.fset.Position(comment.Pos()), text)
		return
	}

	// for /*-style comments, print line by line and let the
	// write function take care of the proper indentation
	lines := split(text)
	stripCommonPrefix(lines)

	// write comment lines, separated by formfeed,
	// without a line break after the last line
	linebreak := formfeeds[0:1]
	pos := p.fset.Position(comment.Pos())
	for i, line := range lines {
		if i > 0 {
			p.write(linebreak)
			pos = p.pos
		}
		if len(line) > 0 {
			p.writeCommentLine(comment, pos, line)
		}
	}
}


// writeCommentSuffix writes a line break after a comment if indicated
// and processes any leftover indentation information. If a line break
// is needed, the kind of break (newline vs formfeed) depends on the
// pending whitespace. writeCommentSuffix returns true if a pending
// formfeed was dropped from the whitespace buffer.
//
func (p *printer) writeCommentSuffix(needsLinebreak bool) (droppedFF bool) {
	for i, ch := range p.buffer {
		switch ch {
		case blank, vtab:
			// ignore trailing whitespace
			p.buffer[i] = ignore
		case indent, unindent:
			// don't loose indentation information
		case newline, formfeed:
			// if we need a line break, keep exactly one
			// but remember if we dropped any formfeeds
			if needsLinebreak {
				needsLinebreak = false
			} else {
				if ch == formfeed {
					droppedFF = true
				}
				p.buffer[i] = ignore
			}
		}
	}
	p.writeWhitespace(len(p.buffer))

	// make sure we have a line break
	if needsLinebreak {
		p.write([]byte{'\n'})
	}

	return
}


// intersperseComments consumes all comments that appear before the next token
// tok and prints it together with the buffered whitespace (i.e., the whitespace
// that needs to be written before the next token). A heuristic is used to mix
// the comments and whitespace. intersperseComments returns true if a pending
// formfeed was dropped from the whitespace buffer.
//
func (p *printer) intersperseComments(next token.Position, tok token.Token) (droppedFF bool) {
	var last *ast.Comment
	for ; p.commentBefore(next); p.cindex++ {
		for _, c := range p.comments[p.cindex].List {
			p.writeCommentPrefix(p.fset.Position(c.Pos()), next, last, tok.IsKeyword())
			p.writeComment(c)
			last = c
		}
	}

	if last != nil {
		if last.Text[1] == '*' && p.fset.Position(last.Pos()).Line == next.Line {
			// the last comment is a /*-style comment and the next item
			// follows on the same line: separate with an extra blank
			p.write([]byte{' '})
		}
		// ensure that there is a line break after a //-style comment,
		// before a closing '}' unless explicitly disabled, or at eof
		needsLinebreak :=
			last.Text[1] == '/' ||
				tok == token.RBRACE && p.mode&noExtraLinebreak == 0 ||
				tok == token.EOF
		return p.writeCommentSuffix(needsLinebreak)
	}

	// no comment was written - we should never reach here since
	// intersperseComments should not be called in that case
	p.internalError("intersperseComments called without pending comments")
	return false
}


// whiteWhitespace writes the first n whitespace entries.
func (p *printer) writeWhitespace(n int) {
	// write entries
	var data [1]byte
	for i := 0; i < n; i++ {
		switch ch := p.buffer[i]; ch {
		case ignore:
			// ignore!
		case indent:
			p.indent++
		case unindent:
			p.indent--
			if p.indent < 0 {
				p.internalError("negative indentation:", p.indent)
				p.indent = 0
			}
		case newline, formfeed:
			// A line break immediately followed by a "correcting"
			// unindent is swapped with the unindent - this permits
			// proper label positioning. If a comment is between
			// the line break and the label, the unindent is not
			// part of the comment whitespace prefix and the comment
			// will be positioned correctly indented.
			if i+1 < n && p.buffer[i+1] == unindent {
				// Use a formfeed to terminate the current section.
				// Otherwise, a long label name on the next line leading
				// to a wide column may increase the indentation column
				// of lines before the label; effectively leading to wrong
				// indentation.
				p.buffer[i], p.buffer[i+1] = unindent, formfeed
				i-- // do it again
				continue
			}
			fallthrough
		default:
			data[0] = byte(ch)
			p.write(data[0:])
		}
	}

	// shift remaining entries down
	i := 0
	for ; n < len(p.buffer); n++ {
		p.buffer[i] = p.buffer[n]
		i++
	}
	p.buffer = p.buffer[0:i]
}


// ----------------------------------------------------------------------------
// Printing interface


func mayCombine(prev token.Token, next byte) (b bool) {
	switch prev {
	case token.INT:
		b = next == '.' // 1.
	case token.ADD:
		b = next == '+' // ++
	case token.SUB:
		b = next == '-' // --
	case token.QUO:
		b = next == '*' // /*
	case token.LSS:
		b = next == '-' || next == '<' // <- or <<
	case token.AND:
		b = next == '&' || next == '^' // && or &^
	}
	return
}


// print prints a list of "items" (roughly corresponding to syntactic
// tokens, but also including whitespace and formatting information).
// It is the only print function that should be called directly from
// any of the AST printing functions in nodes.go.
//
// Whitespace is accumulated until a non-whitespace token appears. Any
// comments that need to appear before that token are printed first,
// taking into account the amount and structure of any pending white-
// space for best comment placement. Then, any leftover whitespace is
// printed, followed by the actual token.
//
func (p *printer) print(args ...interface{}) {
	for _, f := range args {
		next := p.pos // estimated position of next item
		var data []byte
		var tok token.Token

		switch x := f.(type) {
		case pmode:
			// toggle printer mode
			p.mode ^= x
		case whiteSpace:
			if x == ignore {
				// don't add ignore's to the buffer; they
				// may screw up "correcting" unindents (see
				// LabeledStmt)
				break
			}
			i := len(p.buffer)
			if i == cap(p.buffer) {
				// Whitespace sequences are very short so this should
				// never happen. Handle gracefully (but possibly with
				// bad comment placement) if it does happen.
				p.writeWhitespace(i)
				i = 0
			}
			p.buffer = p.buffer[0 : i+1]
			p.buffer[i] = x
		case *ast.Ident:
			data = []byte(x.Name)
			tok = token.IDENT
		case *ast.BasicLit:
			// escape all literals so they pass through unchanged
			// (note that valid Go programs cannot contain
			// tabwriter.Escape bytes since they do not appear in
			// legal UTF-8 sequences)
			data = make([]byte, 0, len(x.Value)+2)
			data = append(data, tabwriter.Escape)
			data = append(data, x.Value...)
			data = append(data, tabwriter.Escape)
			tok = x.Kind
			// If we have a raw string that spans multiple lines and
			// the opening quote (`) is on a line preceded only by
			// indentation, we don't want to write that indentation
			// because the following lines of the raw string are not
			// indented. It's easiest to correct the output at the end
			// via the trimmer (because of the complex handling of
			// white space).
			// Mark multi-line raw strings by replacing the opening
			// quote with esc2 and have the trimmer take care of fixing
			// it up. (Do this _after_ making a copy of data!)
			if data[1] == '`' && bytes.IndexByte(data, '\n') > 0 {
				data[1] = esc2
			}
		case token.Token:
			s := x.String()
			if mayCombine(p.lastTok, s[0]) {
				// the previous and the current token must be
				// separated by a blank otherwise they combine
				// into a different incorrect token sequence
				// (except for token.INT followed by a '.' this
				// should never happen because it is taken care
				// of via binary expression formatting)
				if len(p.buffer) != 0 {
					p.internalError("whitespace buffer not empty")
				}
				p.buffer = p.buffer[0:1]
				p.buffer[0] = ' '
			}
			data = []byte(s)
			tok = x
		case token.Pos:
			if x.IsValid() {
				next = p.fset.Position(x) // accurate position of next item
			}
			tok = p.lastTok
		default:
			fmt.Fprintf(os.Stderr, "print: unsupported argument type %T\n", f)
			panic("go/printer type")
		}
		p.lastTok = tok
		p.pos = next

		if data != nil {
			droppedFF := p.flush(next, tok)

			// intersperse extra newlines if present in the source
			// (don't do this in flush as it will cause extra newlines
			// at the end of a file) - use formfeeds if we dropped one
			// before
			p.writeNewlines(next.Line-p.pos.Line, droppedFF)

			p.writeItem(next, data)
		}
	}
}


// commentBefore returns true iff the current comment occurs
// before the next position in the source code.
//
func (p *printer) commentBefore(next token.Position) bool {
	return p.cindex < len(p.comments) && p.fset.Position(p.comments[p.cindex].List[0].Pos()).Offset < next.Offset
}


// Flush prints any pending comments and whitespace occurring
// textually before the position of the next token tok. Flush
// returns true if a pending formfeed character was dropped
// from the whitespace buffer as a result of interspersing
// comments.
//
func (p *printer) flush(next token.Position, tok token.Token) (droppedFF bool) {
	if p.commentBefore(next) {
		// if there are comments before the next item, intersperse them
		droppedFF = p.intersperseComments(next, tok)
	} else {
		// otherwise, write any leftover whitespace
		p.writeWhitespace(len(p.buffer))
	}
	return
}


// ----------------------------------------------------------------------------
// Trimmer

// A trimmer is an io.Writer filter for stripping tabwriter.Escape
// characters, trailing blanks and tabs, and for converting formfeed
// and vtab characters into newlines and htabs (in case no tabwriter
// is used). Text bracketed by tabwriter.Escape characters is passed
// through unchanged.
//
type trimmer struct {
	output  io.Writer
	state   int
	space   bytes.Buffer
	hasText bool
}


// trimmer is implemented as a state machine.
// It can be in one of the following states:
const (
	inSpace  = iota // inside space
	atEscape        // inside space and the last char was an opening tabwriter.Escape
	inEscape        // inside text bracketed by tabwriter.Escapes
	inText          // inside text
)


var backquote = []byte{'`'}


// Design note: It is tempting to eliminate extra blanks occurring in
//              whitespace in this function as it could simplify some
//              of the blanks logic in the node printing functions.
//              However, this would mess up any formatting done by
//              the tabwriter.

func (p *trimmer) Write(data []byte) (n int, err os.Error) {
	// invariants:
	// p.state == inSpace, atEscape:
	//	p.space is unwritten
	//	p.hasText indicates if there is any text on this line
	// p.state == inEscape, inText:
	//	data[m:n] is unwritten
	m := 0
	var b byte
	for n, b = range data {
		if b == '\v' {
			b = '\t' // convert to htab
		}
		switch p.state {
		case inSpace:
			switch b {
			case '\t', ' ':
				p.space.WriteByte(b) // WriteByte returns no errors
			case '\n', '\f':
				p.space.Reset()                        // discard trailing space
				_, err = p.output.Write(newlines[0:1]) // write newline
				p.hasText = false
			case tabwriter.Escape:
				p.state = atEscape
			default:
				_, err = p.output.Write(p.space.Bytes())
				p.state = inText
				m = n
			}
		case atEscape:
			// discard indentation if we have a multi-line raw string
			// (see printer.print for details)
			if b != esc2 || p.hasText {
				_, err = p.output.Write(p.space.Bytes())
			}
			p.state = inEscape
			m = n
			if b == esc2 {
				_, err = p.output.Write(backquote) // convert back
				m++
			}
		case inEscape:
			if b == tabwriter.Escape {
				_, err = p.output.Write(data[m:n])
				p.state = inSpace
				p.space.Reset()
				p.hasText = true
			}
		case inText:
			switch b {
			case '\t', ' ':
				_, err = p.output.Write(data[m:n])
				p.state = inSpace
				p.space.Reset()
				p.space.WriteByte(b) // WriteByte returns no errors
				p.hasText = true
			case '\n', '\f':
				_, err = p.output.Write(data[m:n])
				p.state = inSpace
				p.space.Reset()
				_, err = p.output.Write(newlines[0:1]) // write newline
				p.hasText = false
			case tabwriter.Escape:
				_, err = p.output.Write(data[m:n])
				p.state = atEscape
				p.space.Reset()
				p.hasText = true
			}
		}
		if err != nil {
			return
		}
	}
	n = len(data)

	switch p.state {
	case inEscape, inText:
		_, err = p.output.Write(data[m:n])
		p.state = inSpace
		p.space.Reset()
		p.hasText = true
	}

	return
}


// ----------------------------------------------------------------------------
// Public interface

// General printing is controlled with these Config.Mode flags.
const (
	RawFormat uint = 1 << iota // do not use a tabwriter; if set, UseSpaces is ignored
	TabIndent                  // use tabs for indentation independent of UseSpaces
	UseSpaces                  // use spaces instead of tabs for alignment
)


// A Config node controls the output of Fprint.
type Config struct {
	Mode     uint // default: 0
	Tabwidth int  // default: 8
}


// Fprint "pretty-prints" an AST node to output and returns the number
// of bytes written and an error (if any) for a given configuration cfg.
// Position information is interpreted relative to the file set fset.
// The node type must be *ast.File, or assignment-compatible to ast.Expr,
// ast.Decl, ast.Spec, or ast.Stmt.
//
func (cfg *Config) Fprint(output io.Writer, fset *token.FileSet, node interface{}) (int, os.Error) {
	// redirect output through a trimmer to eliminate trailing whitespace
	// (Input to a tabwriter must be untrimmed since trailing tabs provide
	// formatting information. The tabwriter could provide trimming
	// functionality but no tabwriter is used when RawFormat is set.)
	output = &trimmer{output: output}

	// setup tabwriter if needed and redirect output
	var tw *tabwriter.Writer
	if cfg.Mode&RawFormat == 0 {
		minwidth := cfg.Tabwidth

		padchar := byte('\t')
		if cfg.Mode&UseSpaces != 0 {
			padchar = ' '
		}

		twmode := tabwriter.DiscardEmptyColumns
		if cfg.Mode&TabIndent != 0 {
			minwidth = 0
			twmode |= tabwriter.TabIndent
		}

		tw = tabwriter.NewWriter(output, minwidth, cfg.Tabwidth, 1, padchar, twmode)
		output = tw
	}

	// setup printer and print node
	var p printer
	p.init(output, cfg, fset)
	go func() {
		switch n := node.(type) {
		case ast.Expr:
			p.nesting = 1
			p.useNodeComments = true
			p.expr(n, ignoreMultiLine)
		case ast.Stmt:
			p.nesting = 1
			p.useNodeComments = true
			// A labeled statement will un-indent to position the
			// label. Set indent to 1 so we don't get indent "underflow".
			if _, labeledStmt := n.(*ast.LabeledStmt); labeledStmt {
				p.indent = 1
			}
			p.stmt(n, false, ignoreMultiLine)
		case ast.Decl:
			p.nesting = 1
			p.useNodeComments = true
			p.decl(n, ignoreMultiLine)
		case ast.Spec:
			p.nesting = 1
			p.useNodeComments = true
			p.spec(n, 1, false, ignoreMultiLine)
		case *ast.File:
			p.nesting = 0
			p.comments = n.Comments
			p.useNodeComments = n.Comments == nil
			p.file(n)
		default:
			p.errors <- fmt.Errorf("printer.Fprint: unsupported node type %T", n)
			runtime.Goexit()
		}
		p.flush(token.Position{Offset: infinity, Line: infinity}, token.EOF)
		p.errors <- nil // no errors
	}()
	err := <-p.errors // wait for completion of goroutine

	// flush tabwriter, if any
	if tw != nil {
		tw.Flush() // ignore errors
	}

	return p.written, err
}


// Fprint "pretty-prints" an AST node to output.
// It calls Config.Fprint with default settings.
//
func Fprint(output io.Writer, fset *token.FileSet, node interface{}) os.Error {
	_, err := (&Config{Tabwidth: 8}).Fprint(output, fset, node) // don't care about number of bytes written
	return err
}
