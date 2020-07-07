package main

import "syscall"
import "os"

// --- libs ---
func fmtSprintf(format string, a []string) string {
	var buf []uint8
	var inPercent bool
	var argIndex int
	var c uint8
	for _, c = range []uint8(format) {
		if inPercent {
			if c == '%' {
				buf = append(buf, c)
			} else {
				var arg = a[argIndex]
				argIndex++
				var s = arg // // p.printArg(arg, c)
				var _c uint8
				for _, _c = range []uint8(s) {
					buf = append(buf, _c)
				}
			}
			inPercent = false
		} else {
			if c == '%' {
				inPercent = true
			} else {
				buf = append(buf, c)
			}
		}
	}

	return string(buf)
}

func fmtPrintf(format string, a ...string) {
	var s = fmtSprintf(format, a)
	syscall.Write(1, []uint8(s))
}


func Itoa(ival int) string {
	var __itoa_buf []uint8 = make([]uint8, 100,100)
	var __itoa_r []uint8 = make([]uint8, 100, 100)

	var next int
	var right int
	var ix = 0
	if ival == 0 {
		return "0"
	}
	var minus bool
	minus = false
	for ix = 0; ival != 0; ix = ix + 1 {
		if ival < 0 {
			ival = -1 * ival
			minus = true
			__itoa_r[0] = '-'
		} else {
			next = ival / 10
			right = ival - next*10
			ival = next
			__itoa_buf[ix] = uint8('0' + right)
		}
	}

	var j int
	var c uint8
	for j = 0; j < ix; j = j + 1 {
		c = __itoa_buf[ix-j-1]
		if minus {
			__itoa_r[j+1] = c
		} else {
			__itoa_r[j] = c
		}
	}

	return string(__itoa_r[0:ix])
}

func inArray(x string, list []string) bool {
	var v string
	for _, v = range list {
		if v == x {
			return true
		}
	}
	return false
}

// --- scanner ---
var scannerSrc []uint8
var scannerCh uint8
var scannerOffset int
var scannerNextOffset int
var scannerInsertSemi bool

func scannerNext() {
	if scannerNextOffset < len(scannerSrc) {
		scannerOffset = scannerNextOffset
		scannerCh = scannerSrc[scannerOffset]
		scannerNextOffset++
	} else {
		scannerOffset = len(scannerSrc)
		scannerCh = 1 //EOF
	}
}

var keywords []string

func scannerInit(src []uint8) {
	// https://golang.org/ref/spec#Keywords
	keywords = []string{
		"break", "default", "func", "interface", "select",
		"case", "defer", "go", "map", "struct",
		"chan", "else", "goto", "package", "switch",
		"const", "fallthrough", "if", "range", "type",
		"continue", "for", "import", "return", "var",
	}
	scannerSrc = src
	scannerOffset = 0
	scannerNextOffset = 0
	scannerInsertSemi = false
	scannerCh = ' '
	fmtPrintf("# src len = %s\n", Itoa(len(scannerSrc)))
	scannerNext()
}

func isLetter(ch uint8) bool {
	if ch == '_' {
		return true
	}
	return ('A' <= ch && ch <= 'Z') || ('a' <= ch && ch <= 'z')
}

func isDecimal(ch uint8) bool {
	return '0' <= ch && ch <= '9'
}

func scannerScanIdentifier() string {
	var offset = scannerOffset
	for isLetter(scannerCh) || isDecimal(scannerCh) {
		scannerNext()
	}
	return string(scannerSrc[offset:scannerOffset])
}

func scannerScanNumber() string {
	var offset = scannerOffset
	for isDecimal(scannerCh) {
		scannerNext()
	}
	return string(scannerSrc[offset:scannerOffset])
}

func scannerScanString() string {
	//fmtPrintf("begin: scannerScanString\n")
	var offset = scannerOffset - 1
	for scannerCh != '"' {
		scannerNext()
	}
	scannerNext() // consume ending '""
	return string(scannerSrc[offset:scannerOffset])
}

func scannerScanChar() string {
	//fmtPrintf("begin: scannerScanString\n")
	var offset = scannerOffset - 1
	for scannerCh != '\'' {
		//fmtPrintf("in loop char:%s\n", string([]uint8{scannerCh}))
		scannerNext()
	}
	scannerNext() // consume ending '""
	return string(scannerSrc[offset:scannerOffset])
}

func scannerrScanComment() string {
	var offset = scannerOffset - 1
	for scannerCh != '\n' {
		scannerNext()
	}
	return string(scannerSrc[offset:scannerOffset])
}

type TokenContainer struct {
	pos int    // what's this ?
	tok string // token.Token
	lit string // raw data
}

// https://golang.org/ref/spec#Tokens
func scannerSkipWhitespace() {
	for scannerCh == ' ' || scannerCh == '\t' || (scannerCh == '\n' && !scannerInsertSemi) || scannerCh == '\r' {
		scannerNext()
	}
}

func scannerScan() *TokenContainer {
	scannerSkipWhitespace()
	var tc = new(TokenContainer)
	var lit string
	var tok string
	var insertSemi bool
	var ch = scannerCh
	if isLetter(ch) {
		lit = scannerScanIdentifier()
		if inArray(lit, keywords) {
			tok = lit
			switch tok {
			case "break", "continue", "fallthrough", "return":
				insertSemi = true
			}
		} else {
			insertSemi = true
			tok = "IDENT"
		}
	} else if isDecimal(ch) {
		insertSemi = true
		lit = scannerScanNumber()
		tok = "INT"
	} else {
		scannerNext()
		switch ch {
		case '\n':
			tok = ";"
			lit = "\n"
			insertSemi = false
		case '"': // double quote
			insertSemi = true
			lit = scannerScanString()
			tok = "STRING"
		case '\'': // single quote
			insertSemi = true
			lit = scannerScanChar()
			tok = "CHAR"
		// https://golang.org/ref/spec#Operators_and_punctuation
		//	+    &     +=    &=     &&    ==    !=    (    )
		//	-    |     -=    |=     ||    <     <=    [    ]
		//  *    ^     *=    ^=     <-    >     >=    {    }
		//	/    <<    /=    <<=    ++    =     :=    ,    ;
		//	%    >>    %=    >>=    --    !     ...   .    :
		//	&^          &^=
		case ':': // :=, :
			if scannerCh == '=' {
				scannerNext()
				tok = ":="
			} else {
				tok = ":"
			}
		case '.': // ..., .
			var peekCh = scannerSrc[scannerNextOffset]
			if scannerCh == '.' && peekCh == '.' {
				scannerNext()
				scannerNext()
				tok = "..."
			} else {
				tok = "."
			}
		case ',':
			tok = ","
		case ';':
			tok = ";"
			lit = ";"
		case '(':
			tok = "("
		case ')':
			insertSemi = true
			tok = ")"
		case '[':
			tok = "["
		case ']':
			insertSemi = true
			tok = "]"
		case '{':
			tok = "{"
		case '}':
			insertSemi = true
			tok = "}"
		case '+': // +=, ++, +
			switch scannerCh {
			case '=':
				scannerNext()
				tok = "+="
			case '+':
				scannerNext()
				tok = "++"
				insertSemi = true
			default:
				tok = "+"
			}
		case '-': // -= --  -
			switch scannerCh {
			case '-':
				scannerNext()
				tok = "--"
				insertSemi = true
			case '=':
				scannerNext()
				tok = "-="
			default:
				tok = "-"
			}
		case '*': // *=  *
			if scannerCh == '=' {
				scannerNext()
				tok = "*="
			} else {
				tok = "*"
			}
		case '/':
			// @TODO block comment
			if scannerCh == '/' {
				lit = scannerrScanComment()
				tok = "COMMENT"
				scannerInsertSemi = false
				return scannerScan()
			} else if scannerCh == '=' {
				tok = "/="
			} else {
				tok = "/"
			}
		case '%': // %= %
			if scannerCh == '=' {
				scannerNext()
				tok = "%="
			} else {
				tok = "%"
			}
		case '^': // ^= ^
			if scannerCh == '=' {
				scannerNext()
				tok = "^="
			} else {
				tok = "^"
			}
		case '<': //  <= <- <<= <<
			switch scannerCh {
			case '-':
				scannerNext()
				tok = "<-"
			case '=':
				scannerNext()
				tok = "<="
			case '<':
				var peekCh = scannerSrc[scannerNextOffset]
				if peekCh == '=' {
					scannerNext()
					scannerNext()
					tok = "<<="
				} else {
					scannerNext()
					tok = "<<"
				}
			default:
				tok = "<"
			}
		case '>': // >= >>= >> >
			switch scannerCh {
			case '=':
				scannerNext()
				tok = ">="
			case '>':
				var peekCh = scannerSrc[scannerNextOffset]
				if peekCh == '=' {
					scannerNext()
					scannerNext()
					tok = ">>="
				} else {
					scannerNext()
					tok = ">>"
				}
			default:
				tok = ">"
			}
		case '=': // == =
			if scannerCh == '=' {
				scannerNext()
				tok = "=="
			} else {
				tok = "="
			}
		case '!': // !=, !
			if scannerCh == '=' {
				scannerNext()
				tok = "!="
			} else {
				tok = "!"
			}
		case '&': // & &= && &^ &^=
			switch scannerCh {
			case '=':
				tok = "&="
			case '&':
				tok = "&&"
			case '^':
				var peekCh = scannerSrc[scannerNextOffset]
				if peekCh == '=' {
					scannerNext()
					scannerNext()
					tok = "&^="
				} else {
					scannerNext()
					tok = "&^"
				}
			default:
				tok = "&"
			}
		case '|': // |= || |
			switch scannerCh {
			case '|':
				scannerNext()
				tok = "||"
			case '=':
				scannerNext()
				tok = "|="
			default:
				tok = "|"
			}
		case 1:
			tok = "EOF"
		default:
			fmtPrintf("unknown char:%s:%s\n", string([]uint8{ch}), Itoa(int(ch)))
			os.Exit(1)
			tok = "UNKNOWN"
		}
	}
	tc.lit = lit
	tc.pos = 0
	tc.tok = tok
	scannerInsertSemi = insertSemi
	return tc
}

// --- ast ---

type astImportSpec struct {
	Path string
}

// Pseudo interface for *ast.Decl
type astDecl struct {
	dtype string
	genDecl *astGenDecl
	funcDecl *astFuncDecl
}

type astFuncDecl struct {
	Name *astIdent
	Sig  *signature
	Body *astBlockStmt
}

type astField struct {
	Name *astIdent
	Type *astExpr
}

type astFieldList struct {
	List []*astField
}

type signature struct {
	params *astFieldList
	results *astFieldList
}

type astStmt struct {
	dtype      string
	DeclStmt   *astDeclStmt
	exprStmt   *astExprStmt
	blockStmt  *astBlockStmt
	assignStmt *astAssignStmt
	returnStmt *astReturnStmt
}

type astDeclStmt struct {
	Decl *astDecl
}

type astExprStmt struct {
	X *astExpr
}

type astBlockStmt struct {
	List []*astStmt
}

type astAssignStmt struct {
	Lhs *astExpr
	Tok string
	Rhs *astExpr
}

type astGenDecl struct {
	Spec *astValueSpec
}

type ObjDecl struct {
	dtype string
	valueSpec *astValueSpec
	funcDecl *astFuncDecl
	field *astField
}

type astValueSpec struct {
	Name *astIdent
	Type *astExpr
	Value *astExpr
}

type astExpr struct {
	dtype        string
	ident        *astIdent
	arrayType    *astArrayType
	basicLit     *astBasicLit
	callExpr     *astCallExpr
	binaryExpr   *astBinaryExpr
	unaryExpr    *astUnaryExpr
	selectorExpr *astSelectorExpr
	indexExpr    *astIndexExpr
}

type astObject struct {
	Kind string
	Name string
	Decl *ObjDecl
	Variable *Variable
}

type astIdent struct {
	Name string
	Obj *astObject
}

type astArrayType struct {
	Len *astExpr
	Elt *astExpr
}

type astBasicLit struct {
	Kind     string   // token.INT, token.CHAR, or token.STRING
	Value    string
}

type astSelectorExpr struct {
	X *astExpr
	Sel *astIdent
}

type astCallExpr struct {
	Fun      *astExpr      // function expression
	Args     []*astExpr    // function arguments; or nil
}

type astUnaryExpr struct {
	X *astExpr
	Op string
}

type astBinaryExpr struct {
	X *astExpr
	Y *astExpr
	Op string
}

type astIndexExpr struct {
	X *astExpr
	Index *astExpr
}

type astScope struct {
	Outer *astScope
	Objects []*objectEntry
}

func astNewScope(outer *astScope) *astScope {
	var r = new(astScope)
	r.Outer = outer
	return r
}

func scopeInsert(s *astScope, obj *astObject) {
	if s == nil {
		panic("[scopeInsert] s sholud not be nil\n")
	}
	var oe = new(objectEntry)
	oe.name = obj.Name
	oe.obj = obj
	s.Objects = append(s.Objects, oe)
}

func scopeLookup(s *astScope, name string) *astObject {
	var oe *objectEntry
	for _, oe =  range s.Objects {
		if oe.name == name {
			return oe.obj
		}
	}
	var r *astObject
	return r
}

// --- parser ---

const O_READONLY int = 0
const FILE_SIZE int = 20000

func readFile(filename string) []uint8 {
	var fd int
	fd, _ = syscall.Open(filename, O_READONLY, 0)
	//fmtPrintf(Itoa(fd))
	//fmtPrintf("\n")
	var buf []uint8 = make([]uint8, FILE_SIZE, FILE_SIZE)
	var n int
	n, _ = syscall.Read(fd, buf)
	//fmtPrintf(Itoa(n))
	var readbytes = buf[0:n]
	return readbytes
}

func readSource(filename string) []uint8 {
	return readFile(filename)
}

func parserInit(src []uint8) {
	scannerInit(src)
	parserNext()
}

type objectEntry struct {
	name string
	obj *astObject
}

var ptok *TokenContainer

var parserTopScope *astScope
var parserPkgScope *astScope

func closeScope() {
	parserTopScope = parserTopScope.Outer
}

func parserNext() {
	//fmtPrintf("parserNext\n")
	ptok = scannerScan()
	if ptok.tok == ";" {
		fmtPrintf("# [parser] looking at : [%s] newline (%s)\n", ptok.tok , Itoa(scannerOffset))
	} else {
		fmtPrintf("# [parser] looking at: [%s] %s (%s)\n", ptok.tok, ptok.lit, Itoa(scannerOffset))
	}
	//fmtPrintf("current ptok: tok=%s, lit=%s\n", ptok.tok, ptok.lit)
}

func parserExpect(tok string, who string) {
	if ptok.tok != tok {
		fmtPrintf("# [%s] %s expected, but got %s\n", who, tok, ptok.tok)
		os.Exit(1)
	}
	fmtPrintf("# [%s] got expected tok %s\n", who, ptok.tok)
	parserNext()
}

func parserExpectSemi(caller string) {
	//fmtPrintf("parserExpectSemi\n")
	if ptok.tok != ")" && ptok.tok != "}" {
		switch ptok.tok {
		case ";":
			parserNext()
		default:
			fmtPrintf("[%s] ; expected, but got %s(%s)\n", caller, ptok.tok, ptok.lit)
			os.Exit(1)
		}
	}
}

func parseIdent() *astIdent {
	var name string
	if ptok.tok == "IDENT" {
		name = ptok.lit
		parserNext()
	} else {
		fmtPrintf("IDENT expected, but got %s\n", ptok.tok)
		os.Exit(1)
	}

	var r = new(astIdent)
	r.Name = name
	return r
}

func parserParseImportDecl() *astImportSpec {
	//fmtPrintf("parserParseImportDecl\n")
	parserExpect("import", "parserParseImportDecl")
	var path = ptok.lit
	parserNext()
	parserExpectSemi("parserParseImportDecl")
	var spec = new(astImportSpec)
	spec.Path = path
	return spec
}

func parseVarType() *astExpr {
	var e = tryIdentOrType()
	if e != nil {
		parserResolve(e.ident, true)
	}
	return e
}

func parseType() *astExpr {
	var typ = tryIdentOrType()
	return typ
}

func eFromArrayType(t *astArrayType) *astExpr {
	var r = new(astExpr)
	r.dtype = "*astArrayType"
	r.arrayType = t
	return r
}

func parseArrayType() *astExpr {
	parserExpect("[", "parseArrayType")
	var ln *astExpr
	if ptok.tok != "]" {
		ln = parseRhs()
	}
	parserExpect("]", "parseArrayTypes")
	var elt = parseType()
	var arrayType = new(astArrayType)
	arrayType.Elt = elt
	arrayType.Len = ln
	return eFromArrayType(arrayType)
}

func tryIdentOrType() *astExpr {
	var typ = new(astExpr)
	fmtPrintf("# debug 1-1\n")
	switch ptok.tok {
	case "IDENT":
		var ident = parseIdent()
		typ.ident = ident
		typ.dtype = "*astIdent"
	case "[":
		typ = parseArrayType()
		return typ
	default:
		fmtPrintf("[tryIdentOrType] TBI:%s\n", ptok.tok)
		os.Exit(1)
	}
	return typ
}


func parseParameterList(scope *astScope) []*astField {
	var params []*astField
	//var list []*astExpr
	for ptok.tok != ")" {
		var field = new(astField)
		var ident = parseIdent()
//		fmtPrintf("debug 1\n")
		var typ = parseVarType()
//		fmtPrintf("debug 2\n")
		field.Name = ident
		field.Type = typ
//		fmtPrintf("debug 3\n")
		fmtPrintf("# [parser] parseParameterList: Field %s %s\n", field.Name.Name, field.Type.dtype)
//		fmtPrintf("debug 4\n")
		params = append(params, field)
		if parserTopScope == nil {
			panic("parserTopScope should not be nil")
		}
		declareField(field, scope, "Var", ident)
		if ptok.tok == "," {
			parserNext()
			continue
		} else if ptok.tok == ")" {
			break
		} else {
			fmtPrintf("[parseParameterList] Syntax Error\n")
			os.Exit(1)
		}
	}

	return params
}

func parseParameters(scope *astScope) *astFieldList {
	var params []*astField
	parserExpect("(", "parseParameters")
	if ptok.tok != ")" {
		params = parseParameterList(scope)
	}
	parserExpect(")", "parseParameters")
	var afl = new(astFieldList)
	afl.List = params
	return afl
}

func parserResult(scope *astScope) *astFieldList {
	var r = new(astFieldList)
	if ptok.tok == "{" {
		r = nil
		return r
	}
	var typ = parseType()
	var field = new(astField)
	field.Type = typ
	r.List = append(r.List, field)
	return r
}

func parseSignature(scope *astScope) *signature {
	var params *astFieldList
	var results *astFieldList
	params = parseParameters(scope)
	results = parserResult(scope)
	var sig = new(signature)
	sig.params = params
	sig.results = results
	return sig
}

func declareField(decl *astField, scope *astScope, kind string, ident *astIdent) {
	// delcare
	var obj = new(astObject)
	var objDecl = new(ObjDecl)
	objDecl.dtype = "*astField"
	objDecl.field = decl
	obj.Decl = objDecl
	obj.Name = ident.Name
	obj.Kind = kind
	ident.Obj = obj

	// scope insert
	if ident.Name != "_" {
		scopeInsert(scope, obj)
	}
}

func declare(objDecl *ObjDecl, scope *astScope, kind string, ident *astIdent) {
	fmtPrintf("# [declare] ident %s\n", ident.Name)

	var obj = new(astObject) //valSpec.Name.Obj
	obj.Decl = objDecl
	obj.Name = ident.Name
	obj.Kind = kind
	ident.Obj = obj

	// scope insert
	if ident.Name != "_" {
		scopeInsert(scope, obj)
	}
	fmtPrintf("# [declare] end\n")

}

var parserUnresolved []*astIdent

func parserResolveEx(ident *astIdent) {
	parserResolve(ident, true)
}
func parserResolve(ident *astIdent, collectUnresolved bool) {
	if ident.Name == "-" {
		return
	}

	var s *astScope
	for s = parserTopScope; s != nil; s = s.Outer {
		var obj = scopeLookup(s, ident.Name)
		if obj != nil {
			ident.Obj = obj
			return
		}
	}

	if collectUnresolved {
		parserUnresolved = append(parserUnresolved, ident)
		fmtPrintf("# collected unresolved ident %s\n", ident.Name)
	}
}

func parseOperand() *astExpr {
	switch ptok.tok {
	case "IDENT":
		var eIdent = new(astExpr)
		eIdent.dtype = "*astIdent"
		var ident  =  parseIdent()
		eIdent.ident = ident
		parserResolve(ident, true)
		return eIdent
	case "INT":
		var basicLit = new(astBasicLit)
		basicLit.Kind = "INT"
		basicLit.Value = ptok.lit
		var r = new(astExpr)
		r.dtype = "*astBasicLit"
		r.basicLit = basicLit
		parserNext()
		return r
	case "STRING":
		var basicLit = new(astBasicLit)
		basicLit.Kind = "STRING"
		basicLit.Value = ptok.lit
		var r = new(astExpr)
		r.dtype = "*astBasicLit"
		r.basicLit = basicLit
		parserNext()
		return r
	case "CHAR":
		var basicLit = new(astBasicLit)
		basicLit.Kind = "CHAR"
		basicLit.Value = ptok.lit
		var r = new(astExpr)
		r.dtype = "*astBasicLit"
		r.basicLit = basicLit
		parserNext()
		return r
	}

	var typ = tryIdentOrType()
	if typ == nil {
		fmtPrintf("# [parseOperand] typ should not be nil\n")
		os.Exit(1)
	}

	return typ
}

func parseCallExpr(fn *astExpr) *astCallExpr {
	parserExpect("(", "parseCallExpr")
	var callExpr = new(astCallExpr)
	callExpr.Fun = fn
	fmtPrintf("# [parsePrimaryExpr] ptok.tok=%s\n", ptok.tok)
	var list []*astExpr
	for ptok.tok != ")" {
		var arg = parseExpr()
		list = append(list, arg)
		if ptok.tok == "," {
			parserNext()
		} else if ptok.tok == ")" {
			break
		}
	}
	parserExpect(")", "parsePrimaryExpr") // consume ")"
	callExpr.Args = list
	return callExpr
}

func parsePrimaryExpr() *astExpr {
	fmtPrintf("#   begin parsePrimaryExpr()\n")
	var x = parseOperand()
	var r = new(astExpr)
	switch ptok.tok {
	case ".":
		parserNext() // consume "."
		if ptok.tok != "IDENT" {
			fmtPrintf("tok should be IDENT")
			os.Exit(1)
		}
		// Assume CallExpr
		var secondIdent = parseIdent()
		if ptok.tok == "(" {
			var fn = new(astExpr)
			fn.dtype = "*astSelectorExpr"
			fn.selectorExpr = new(astSelectorExpr)
			fn.selectorExpr.X = x
			fn.selectorExpr.Sel = secondIdent
			// string = x.ident.Name + "." + secondIdent
			r.dtype = "*astCallExpr"
			r.callExpr = parseCallExpr(fn)
			fmtPrintf("# [parsePrimaryExpr] 741 ptok.tok=%s\n", ptok.tok)
		} else {
			fmtPrintf("#   end parsePrimaryExpr()\n")
			return x
		}
	case "(":
		r.dtype = "*astCallExpr"
		r.callExpr = parseCallExpr(x)
		fmtPrintf("# [parsePrimaryExpr] 741 ptok.tok=%s\n", ptok.tok)
	case "[":
		parserResolveEx(x.ident)
		x = parseIndexOrSlice(x)
		return x
	default:
		fmtPrintf("#   end parsePrimaryExpr()\n")
		return x
	}
	fmtPrintf("#   end parsePrimaryExpr()\n")
	return r

}

func parseIndexOrSlice(x *astExpr) *astExpr {
	parserExpect("[", "parseIndexOrSlice")
	var index0 *astExpr
	if ptok.tok != ":" {
		index0 = parseRhs()
	}

	parserExpect("]", "parseIndexOrSlice")
	var indexExpr = new(astIndexExpr)
	indexExpr.X = x
	indexExpr.Index = index0
	var r = new(astExpr)
	r.dtype = "*astIndexExpr"
	r.indexExpr = indexExpr
	return r
}

func parseUnaryExpr() *astExpr {
	var r *astExpr
	fmtPrintf("#   begin parseUnaryExpr()\n")
	switch ptok.tok {
	case "+","-", "!":
		var tok = ptok.tok
		parserNext()
		var x = parseUnaryExpr()
		r = new(astExpr)
		r.dtype = "*astUnaryExpr"
		r.unaryExpr = new(astUnaryExpr)
		r.unaryExpr.Op = tok
		r.unaryExpr.X = x
		return r
	}
	r  = parsePrimaryExpr()
	fmtPrintf("#   end parseUnaryExpr()\n")
	return r
}

const LowestPrec int = 0

func precedence(op string) int {
	switch op {
	case "||":
		return 1
	case "&&":
		return 2
	case "==", "!=", "<", "<=", ">", ">=":
		return 3
	case "+", "-":
		return 4
	case "*", "/", "%":
		return 5
	default:
		return 0
	}
	return 0
}

func parseBinaryExpr(prec1 int) *astExpr {
	fmtPrintf("#   begin parseBinaryExpr() prec1=%s\n", Itoa(prec1))
	var x = parseUnaryExpr()
	var oprec int
	for {
		var op = ptok.tok
		oprec  = precedence(op)
		fmtPrintf("# oprec %s\n", Itoa(oprec))
		fmtPrintf("# precedence \"%s\" %s < %s\n", op, Itoa(oprec) , Itoa(prec1))
		if oprec < prec1 {
			fmtPrintf("#   end parseBinaryExpr() (NonBinary)\n")
			return x
		}
		parserExpect(op, "parseBinaryExpr")
		var y = parseBinaryExpr(oprec+1)
		var binaryExpr = new(astBinaryExpr)
		binaryExpr.X = x
		binaryExpr.Y = y
		binaryExpr.Op = op
		var r = new(astExpr)
		r.dtype = "*astBinaryExpr"
		r.binaryExpr = binaryExpr
		x = r
	}
	fmtPrintf("#   end parseBinaryExpr()\n")
	return x
}

func parseExpr() *astExpr {
	fmtPrintf("#   begin parseExpr()\n")
	var e = parseBinaryExpr(1)
	fmtPrintf("#   end parseExpr()\n")
	return e
}

func parseRhs() *astExpr {
	var x = parseExpr()
	return x
}

func nop() {
}

func nop1() {
}

func parseStmt() *astStmt {
	fmtPrintf("# = begin parseStmt()\n")
	var s = new(astStmt)
	switch ptok.tok {
	case "var":
		var genDecl = parseDecl("var")
		s.dtype = "*astDeclStmt"
		s.DeclStmt = new(astDeclStmt)
		var decl = new(astDecl)
		decl.dtype = "*astGenDecl"
		decl.genDecl = genDecl
		s.DeclStmt.Decl = decl
		fmtPrintf("# = end parseStmt()\n")
		return s
	case "IDENT":
		fmtPrintf("# [parseStmt] is IDENT:%s\n",ptok.lit)
		var x = parseExpr()
		var stok = ptok.tok
		switch stok {
		case "=":
			parserNext() // consume =
			//fmtPrintf("# [parseStmt] ERROR:%s\n",stok)
			//os.Exit(1)
			var y = parseExpr() // rhs
			var as = new(astAssignStmt)
			as.Tok = "="
			as.Lhs = x
			as.Rhs = y
			s.dtype = "*astAssignStmt"
			s.assignStmt = as
			parserExpectSemi("parseStmt:IDENT")
			fmtPrintf("# = end parseStmt()\n")
			return s
		case ";":
			s.dtype = "*astExprStmt"
			var exprStmt = new(astExprStmt)
			exprStmt.X = x
			s.exprStmt = exprStmt
			parserExpectSemi("parseStmt:,")
			fmtPrintf("# = end parseStmt()\n")
			return s
		default:
			fmtPrintf("parseStmt:Ident:TBI:%s\n", ptok.tok)
			os.Exit(1)
		}
	case "return":
		var s = parseReturnStmt()
		return s
	default:
		fmtPrintf("parseStmt:TBI 3:%s\n", ptok.tok)
		os.Exit(1)
	}
	fmtPrintf("# = end parseStmt()\n")
	return s
}

func parseExprList() []*astExpr {
	var list []*astExpr
	var e = parseExpr()
	list = append(list, e)
	for ptok.tok == "," {
		parserNext() // consume ","
		e = parseExpr()
		list = append(list, e)
	}

	return list
}

func parseRhsList() []*astExpr {
	var list = parseExprList()
	return list
}

type astReturnStmt struct {
	Results []*astExpr
}

func parseReturnStmt() *astStmt {
	parserExpect("return", "parseReturnStmt")
	var x []*astExpr
	if ptok.tok != ";" && ptok.tok != "}" {
		x = parseRhsList()
	}
	parserExpectSemi("parseReturnStmt")
	var returnStmt = new(astReturnStmt)
	returnStmt.Results = x
	var r = new(astStmt)
	r.dtype = "*astReturnStmt"
	r.returnStmt = returnStmt
	return r
}

func parseStmtList() []*astStmt {
	var list []*astStmt
	for ptok.tok != "}" {
		if ptok.tok == "EOF" {
			fmtPrintf("#[parseStmtList] unexpected EOF\n")
			os.Exit(1)
		}
		var stmt = parseStmt()
		list = append(list, stmt)
	}
	return list
}

func parseBody(scope *astScope) *astBlockStmt {
	parserExpect("{", "parseBody")
	parserTopScope = scope
	var list  []*astStmt
	fmtPrintf("# begin parseStmtList()\n")
	list = parseStmtList()
	fmtPrintf("# end parseStmtList()\n")

	closeScope()
	parserExpect("}", "parseBody")
	var r = new(astBlockStmt)
	r.List = list
	return r
}

func parseDecl(keyword string) *astGenDecl {
	var r *astGenDecl
	switch ptok.tok {
	case "var":
		parserExpect(keyword, "parseDecl:var:1")
		var ident = parseIdent()
		var typ = parseType()
		var value *astExpr
		if ptok.tok == "=" {
			parserNext()
			value = parseExpr()
		}
		parserExpectSemi("parseDecl:var:2")
		var spec = new(astValueSpec)
		spec.Name = ident
		spec.Type = typ
		spec.Value = value
		var objDecl = new(ObjDecl)
		objDecl.dtype = "*astValueSpec"
		objDecl.valueSpec = spec
		declare(objDecl, parserTopScope, "Var", ident)
		r = new(astGenDecl)
		r.Spec = spec
		return r
	default:
		fmtPrintf("[parseDecl] TBI\n")
		os.Exit(1)
	}
	return r
}

type astSpec struct {
	dtype string
	valueSpec *astValueSpec
}

func parserParseValueSpec() *astSpec {
	fmtPrintf("# [parserParseValueSpec] start\n")
	parserExpect("var", "parserParseValueSpec")
	var ident = parseIdent()
	fmtPrintf("# [parserParseValueSpec] ident = %s\n", ident.Name)
	var typ = parseType()
	parserExpectSemi("parserParseValueSpec")
	var spec = new(astValueSpec)
	spec.Name = ident
	spec.Type = typ
	var r = new(astSpec)
	r.dtype = "*astValueSpec"
	r.valueSpec = spec
	var objDecl = new(ObjDecl)
	objDecl.dtype = "*astValueSpec"
	objDecl.valueSpec = spec
	declare(objDecl, parserTopScope, "Var", ident)
	fmtPrintf("# [parserParseValueSpec] end\n")
	return r
}

func parserParseFuncDecl() *astDecl {
	parserExpect("func", "parserParseFuncDecl")
	var scope = astNewScope(parserTopScope) // function scope

	var ident = parseIdent()
	var sig = parseSignature(scope)
	if sig.results == nil {
		fmtPrintf("# [parserParseFuncDecl] %s sig.results is nil\n", ident.Name)
	} else {
		fmtPrintf("# [parserParseFuncDecl] %s sig.results.List = %s\n", ident.Name, Itoa(len(sig.results.List)))
	}
	var body *astBlockStmt
	if ptok.tok == "{" {
		fmtPrintf("# begin parseBody()\n")
		body = parseBody(scope)
		fmtPrintf("# end parseBody()\n")
		parserExpectSemi("parserParseFuncDecl")
	}
	var decl = new(astDecl)
	decl.dtype = "*astFuncDecl"
	var funcDecl = new(astFuncDecl)
	decl.funcDecl = funcDecl
	decl.funcDecl.Name = ident
	decl.funcDecl.Sig = sig
	decl.funcDecl.Body = body
	var objDecl = new(ObjDecl)
	objDecl.dtype = "*astFuncDecl"
	objDecl.funcDecl = funcDecl
	declare(objDecl, parserPkgScope, "Fun", ident)
	return decl
}

func parserParseFile() *astFile {
	// expect "package" keyword
	parserExpect("package", "parserParseFile")
	parserUnresolved = nil
	var ident = parseIdent()
	var packageName = ident.Name
	parserExpectSemi("parserParseFile")

	parserTopScope = new(astScope) // open scope
	parserPkgScope = parserTopScope

	for ptok.tok == "import" {
		parserParseImportDecl()
	}

	fmtPrintf("#\n")
	fmtPrintf("# [parser] Parsing Top level decls\n")
	var decls []*astDecl
	var decl *astDecl

	for ptok.tok != "EOF" {
		switch ptok.tok {
		case "var":
			var spec = parserParseValueSpec()
			fmtPrintf("# [parserParseFile] debug 1\n")
			var genDecl = new(astGenDecl)
			fmtPrintf("# [parserParseFile] debug 2\n")
			genDecl.Spec = spec.valueSpec
			fmtPrintf("# [parserParseFile] debug 3\n")
			decl = new(astDecl)
			decl.dtype = "*astGenDecl"
			decl.genDecl = genDecl
			fmtPrintf("# [parserParseFile] debug 9\n")
		case "func":
			decl  = parserParseFuncDecl()
			fmtPrintf("# func decl parsed:%s\n", decl.funcDecl.Name.Name)
		default:
			fmtPrintf("# parserParseFile:TBI:%s\n", ptok.tok)
			os.Exit(1)
		}
		decls = append(decls, decl)
	}

	parserTopScope = nil

	var unresolved []*astIdent
	var idnt *astIdent
	fmtPrintf("# [parserParseFile] resolving parserUnresolved (n=%s)\n", Itoa(len(parserUnresolved)))
	for _, idnt = range parserUnresolved {
		fmtPrintf("# [parserParseFile] resolving ident %s \n", idnt.Name)
		var obj *astObject = scopeLookup(parserPkgScope ,idnt.Name)
		if obj != nil {
			fmtPrintf("# [parserParseFile] obj matched %s \n", obj.Name)
			idnt.Obj = obj
		} else {
			unresolved = append(unresolved, idnt)
		}
	}
	fmtPrintf("# [parserParseFile] Unresolved (n=%s)\n", Itoa(len(unresolved)))

	var f = new(astFile)
	f.Name = packageName
	f.Decls = decls
	f.Unresolved = unresolved
	return f
}


func parseFile(filename string) *astFile {
	var text = readSource(filename)
	parserInit(text)
	return parserParseFile()
}

// --- codegen ---
func emitPopBool(comment string) {
	fmtPrintf("  popq %%rax # result of %s\n", comment)
}

func emitPopAddress(comment string) {
	fmtPrintf("  popq %%rax # address of %s\n", comment)
}

func emitPopString() {
	fmtPrintf("  popq %%rax # string.ptr\n")
	fmtPrintf("  popq %%rcx # string.len\n")
}

func emitPopSlice() {
	fmtPrintf("  popq %%rax # slice.ptr\n")
	fmtPrintf("  popq %%rcx # slice.len\n")
	fmtPrintf("  popq %%rdx # slice.cap\n")
}

func emitPushStackTop(condType *Type, comment string) {
	switch kind(condType) {
	case T_STRING:
		fmtPrintf("  movq 8(%%rsp), %%rcx # copy str.len from stack top (%s)\n", comment)
		fmtPrintf("  movq 0(%%rsp), %%rax # copy str.ptr from stack top (%s)\n", comment)
		fmtPrintf("  pushq %%rcx # str.len\n")
		fmtPrintf("  pushq %%rax # str.ptr\n")
	case T_POINTER, T_UINTPTR, T_BOOL, T_INT, T_UINT8, T_UINT16:
		fmtPrintf("  movq (%%rsp), %%rax # copy stack top value (%s) \n", comment)
		fmtPrintf("  pushq %%rax\n")
	default:
		throw(kind(condType))
	}
}

func emitRevertStackPointer(size int) {
	fmtPrintf("  addq $%s, %%rsp # revert stack pointer\n", Itoa(size))
}

func emitAddConst(addValue int, comment string) {
	fmtPrintf("  # Add const: %s\n", comment)
	fmtPrintf("  popq %%rax\n")
	fmtPrintf("  addq $%s, %%rax\n", Itoa(addValue))
	fmtPrintf("  pushq %%rax\n")
}

func getPushSizeOfType(t *Type) int {
	switch kind(t) {
	case T_SLICE:
		return sliceSize
	case T_STRING:
		return stringSize
	case T_UINT8, T_UINT16, T_INT, T_BOOL:
		return intSize
	case T_UINTPTR, T_POINTER:
		return ptrSize
	case T_ARRAY, T_STRUCT:
		return ptrSize
	default:
		throw(kind(t))
	}
	throw(kind(t))
	return 0
}

func getElementTypeOfListType(t *Type) *Type {
	switch kind(t) {
	case T_SLICE, T_ARRAY:
		var arrayType =  t.e.arrayType
		if arrayType == nil {
			panic("[getElementTypeOfListType] should not be nil")
		}
		return e2t(arrayType.Elt)
	case T_STRING:
		return tUint8
	default:
		panic("[getElementTypeOfListType] TBI kind=" +  kind(t))
	}
	var r *Type
	return r
}

const sliceSize int = 24
const stringSize int = 16
const intSize int = 8
const ptrSize int = 8

func throw(s string) {
	syscall.Write(2, []uint8(s))
}

//type localoffsetint int //@TODO
var localoffset int

type Variable struct {
	name         string
	isGlobal     bool
	globalSymbol string
	localOffset  int
}

func walkStmt(stmt *astStmt) {
	fmtPrintf("# [walkStmt] begin dtype=%s\n", stmt.dtype)
	switch stmt.dtype {
	case "*astDeclStmt":
		fmtPrintf("# [walkStmt] *ast.DeclStmt\n")
		var declStmt = stmt.DeclStmt
		if declStmt.Decl == nil {
			fmtPrintf("[walkStmt] ERROR\n")
			os.Exit(1)
		}
		var dcl = declStmt.Decl
		if dcl.dtype != "*astGenDecl" {
			panic("[walkStmt][dcl.dtype] internal error")
		}

		var valSpec = dcl.genDecl.Spec
		var typ = valSpec.Type // ident "int"
		fmtPrintf("# [walkStmt] valSpec Name=%s, Type=%s\n",
			valSpec.Name.Name, typ.dtype)
		var t = e2t(typ)
		var sizeOfType = getSizeOfType(t)
		localoffset = localoffset - sizeOfType

		valSpec.Name.Obj.Variable = newLocalVariable(valSpec.Name.Name, localoffset)
		fmtPrintf("# var %s offset = %d\n", valSpec.Name.Obj.Name,
			Itoa(valSpec.Name.Obj.Variable.localOffset))
		if valSpec.Value != nil {
			walkExpr(valSpec.Value)
		}
	case "*astAssignStmt":
		var rhs = stmt.assignStmt.Rhs
		walkExpr(rhs)
	default:

	}
}

func panic(x string) {
	fmtPrintf(x+"\n")
	os.Exit(1)
}

var stringLiterals []*stringLiteralsContainer

type stringLiteralsContainer struct {
	lit *astBasicLit
	sl  *sliteral
}

var stringIndex int


func getStringLiteral(lit *astBasicLit) *sliteral {
	var container *stringLiteralsContainer
	for _, container = range stringLiterals {
		if container.lit == lit {
			return container.sl
		}
	}

	fmtPrintf("[getStringLiteral] string literal not found: %s\n" , lit.Value)
	os.Exit(1)
	var r *sliteral
	return r
}

func registerStringLiteral(lit *astBasicLit) {
	fmtPrintf("# [registerStringLiteral] begin\n")

	var pkgName = "main"
	if pkgName == "" {
		panic("no pkgName")
	}

	var strlen int
	var c uint8
	var vl = []uint8(lit.Value)
	for _, c = range vl {
		if c != '\\' {
			strlen++
		}
	}

	var label = fmtSprintf(".%s.S%d", []string{pkgName, Itoa(stringIndex)})
	stringIndex++

	var sl = new(sliteral)
	sl.label = label
	sl.strlen = strlen - 2
	sl.value = lit.Value
	fmtPrintf("# [registerStringLiteral] label=%s, strlen=%s\n", sl.label, Itoa(sl.strlen))
	var cont = new(stringLiteralsContainer)
	cont.sl = sl
	cont.lit = lit
	stringLiterals = append(stringLiterals, cont)
}

func walkExpr(expr *astExpr) {
	fmtPrintf("# [walkExpr] dtype=%s\n", expr.dtype)
	switch expr.dtype {
	case "*astIdent":
		// what to do ?
	case "*astCallExpr":
		var arg *astExpr
		for _, arg = range expr.callExpr.Args {
			walkExpr(arg)
		}
	case "*astBasicLit":
		switch expr.basicLit.Kind {
		case "STRING":
			registerStringLiteral(expr.basicLit)
		}
	case "*astUnaryExpr":
		walkExpr(expr.unaryExpr.X)
	case "*astBinaryExpr":
		walkExpr(expr.binaryExpr.X)
		walkExpr(expr.binaryExpr.Y)
	default:
		panic("[walkExpr] TBI:" + expr.dtype)
	}
}


func newGlobalVariable(name string) *Variable {
	var vr = new(Variable)
	vr.name = name
	vr.isGlobal = true
	vr.globalSymbol = name
	return vr
}

func newLocalVariable(name string, localoffset int) *Variable {
	var vr = new(Variable)
	vr.name = name
	vr.isGlobal = false
	vr.localOffset = localoffset
	return vr
}

var pkgName string

func getSizeOfType(t *Type) int {
	switch kind(t) {
	case T_SLICE:
		return sliceSize
	case T_STRING:
		return 16
	case T_INT:
		return 8
	case T_UINT8:
		return 1
	default:
		fmtPrintf("[getSizeOfType] TBI:%s\n", kind(t))
		os.Exit(1)
	}
	return 0
}

func e2t(typeExpr *astExpr) *Type {
	if typeExpr == nil {
		panic("nil is not allowed")
	}
	var r = new(Type)
	r.e = typeExpr
	return r
}

var gString *astObject
var gInt *astObject
var gUint8 *astObject
var gUint16 *astObject
var gUintptr *astObject
var gBool *astObject

func semanticAnalyze(file *astFile) string {
	// create universe scope
	var universe = new(astScope)
	// inject predeclared identifers
	scopeInsert(universe, gInt)
	scopeInsert(universe, gUint8)
	scopeInsert(universe, gString)

	// @FIXME package names should be be in universe
	var pkgOs = new(astObject)
	pkgOs.Kind = "Pkg"
	pkgOs.Name = "os"
	scopeInsert(universe, pkgOs)

	var pkgSyscall= new(astObject)
	pkgSyscall.Kind = "Pkg"
	pkgSyscall.Name = "syscall"
	scopeInsert(universe, pkgSyscall)

	pkgName = file.Name

	var unresolved []*astIdent
	var ident *astIdent
	fmtPrintf("# [SEMA] resolving file.Unresolved (n=%s)\n", Itoa(len(file.Unresolved)))
	for _, ident = range file.Unresolved {
		fmtPrintf("# [SEMA] resolving ident %s \n", ident.Name)
		var obj *astObject = scopeLookup(universe,ident.Name)
		if obj != nil {
			fmtPrintf("# [SEMA] obj matched %s \n", obj.Name)
			ident.Obj = obj
		} else {
			panic("[semanticAnalyze] Unresolved : " + ident.Name)
			unresolved = append(unresolved, ident)
		}
	}

	var decl *astDecl
	for _, decl = range file.Decls {
		switch decl.dtype {
		case "*astGenDecl":
			var genDecl = decl.genDecl
			var valSpec = genDecl.Spec
			var nameIdent = valSpec.Name
			nameIdent.Obj.Variable = newGlobalVariable(nameIdent.Obj.Name)
			globalVars = append(globalVars, valSpec)
		case "*astFuncDecl":
			var funcDecl = decl.funcDecl
			fmtPrintf("# [sema] == astFuncDecl %s ==\n", funcDecl.Name.Name)
			localoffset  = 0
			var paramoffset = 16
			var field *astField
			for _, field = range funcDecl.Sig.params.List {
				var obj = field.Name.Obj
				obj.Variable = newLocalVariable(obj.Name, paramoffset)
				var varSize = getSizeOfType(e2t(field.Type))
				paramoffset = paramoffset + varSize
				fmtPrintf("# field.Name.Obj.Name=%s\n", obj.Name)
				//fmtPrintf("#   field.Type=%#v\n", field.Type)
			}
			var stmt *astStmt
			for _, stmt = range funcDecl.Body.List {
				walkStmt(stmt)
			}
			var fnc = new(Func)
			fnc.name = funcDecl.Name.Name
			fnc.Body = funcDecl.Body
			fnc.localarea = localoffset
			globalFuncs = append(globalFuncs, fnc)
		default:
			fmtPrintf("[semanticAnalyze] TBI: %s\n", decl.dtype)
			os.Exit(1)
		}
	}

	if len(stringLiterals) == 0 {
		panic("# [sema] stringLiterals is empty\n")
	}

	return pkgName
}


var T_STRING string
var T_SLICE string
var T_BOOL string
var T_INT string
var T_UINT8 string
var T_UINT16 string
var T_UINTPTR string
var T_ARRAY string
var T_STRUCT string
var T_POINTER string



func emitGlobalVariable(name *astIdent, t *Type, val *astExpr) {
	var typeKind = kind(t)
	fmtPrintf("%s: # T %s \n", name.Name, typeKind)
	switch typeKind {
	case T_STRING:
		fmtPrintf("  .quad 0\n")
		fmtPrintf("  .quad 0\n")
	case T_UINTPTR:
		fmtPrintf("  .quad 0\n")
	case T_BOOL:
		fmtPrintf("  .quad 0 # bool zero value\n") // @TODO
	case T_INT:
		fmtPrintf("  .quad 0\n")
	case T_UINT8:
		fmtPrintf("  .byte 0\n")
	case T_UINT16:
		fmtPrintf("  .word 0\n")
	case T_ARRAY:
		if val != nil {
			panic("[emitGlobalVariable] TBI")
		}
		if t.e.dtype != "*astArrayType" {
			panic("[emitGlobalVariable] Unexpected type:" + t.e.dtype)
		}
		var arrayType = t.e.arrayType
		if arrayType.Len == nil {
			panic("[emitGlobalVariable] global slice is not supported")
		}
		if arrayType.Len.dtype != "*astBasicLit" {
			panic("[emitGlobalVariable] shoulbe basic literal")
		}
		var basicLit = arrayType.Len.basicLit
		var v = basicLit.Value[0]
		var length = int(v - '0')
		fmtPrintf("# [emitGlobalVariable] array length uint8=%d\n" , Itoa(length))
		var zeroValue string
		var kind string = kind(e2t(arrayType.Elt))
		switch kind {
		case T_UINT8:
			zeroValue = "  .byte 0 # uint8 zero value\n"
		default:
			panic("[emitGlobalVariable] Unexpected kind:" + kind)
		}

		fmtPrintf("# -- debug 8 \n")
		var i int
		for i = 0; i< length ; i++ {
			fmtPrintf(zeroValue)
		}
		fmtPrintf("# -- debug 9 \n")
	default:
		panic("[emitGlobalVariable] TBI:kind=" + typeKind)
	}
}

func emitData(pkgName string) {
	fmtPrintf(".data\n")
	fmtPrintf("# string literals len = %s\n", Itoa(len(stringLiterals)))
	var con *stringLiteralsContainer
	for _, con = range stringLiterals {
		fmtPrintf("# string literals\n")
		fmtPrintf("%s:\n", con.sl.label)
		fmtPrintf("  .string %s\n", con.sl.value)
	}

	fmtPrintf("# ===== Global Variables =====\n")

	var spec *astValueSpec
	for _, spec = range globalVars {
		var t *Type
		if spec.Type != nil {
			t = e2t(spec.Type)
		}
		emitGlobalVariable(spec.Name, t, spec.Value)
	}

	fmtPrintf("# ==============================\n")
}

func emitZeroValue(t *Type) {
	switch kind(t) {
	case T_INT:
		fmtPrintf("  pushq $0 # %s zero value\n", kind(t))
	default:
		panic("[emitZeroValue] TBI:" + kind(t))
	}
}

type Arg struct {
	e      *astExpr
	t      *Type // expected type
	offset int
}

func emitCallNonDecl(symbol string, eArgs []*astExpr) {
	var args []*Arg
	var eArg *astExpr
	for _, eArg = range eArgs {
		var arg = new(Arg)
		arg.e = eArg
		arg.t = nil
		args = append(args, arg)
	}
	emitCall(symbol, args)
}

func emitCall(symbol string, args []*Arg) {
	var pushed int = 0
	//var arg *astExpr
	var i int
	for i=len(args)-1;i>=0;i-- {
		emitExpr(args[i].e)
		var typ = getTypeOfExpr(args[i].e)
		var size = getSizeOfType(typ)
		pushed = pushed + size
	}
	fmtPrintf("  callq %s\n", symbol)
	fmtPrintf("  addq $%s, %%rsp # revert \n", Itoa(pushed))
}

func emitExpr(e *astExpr) {
	fmtPrintf("# [emitExpr] dtype=%s\n", e.dtype)
	switch e.dtype {
	case "*astIdent":
		var ident = e.ident
		if ident.Obj == nil {
			fmtPrintf("[emitExpr] ident %s is unresolved\n", ident.Name)
			os.Exit(1)
		}
		switch ident.Obj.Kind {
		case "Var":
			emitAddr(e)
			var t = getTypeOfExpr(e)
			emitLoad(t)
		default:
			fmtPrintf("unknown Kind=%s\n", ident.Obj.Kind)
			os.Exit(1)
		}
	case "*astIndexExpr":
		emitAddr(e)
		emitLoad(getTypeOfExpr(e))
	case "*astBasicLit":
		fmtPrintf("# basicLit.Kind = %s \n", e.basicLit.Kind)
		switch e.basicLit.Kind {
		case "INT":
			fmtPrintf("  pushq $%s # \n", e.basicLit.Value)
		case "STRING":
			var sl = getStringLiteral(e.basicLit)
			if sl.strlen == 0 {
				// zero value
				//emitZeroValue(tString)
				fmtPrintf("[emitExpr] TBI: empty string literal\n")
				os.Exit(1)
			} else {
				fmtPrintf("  pushq $%d # str len\n", Itoa(sl.strlen))
				fmtPrintf("  leaq %s, %%rax # str ptr\n", sl.label)
				fmtPrintf("  pushq %%rax # str ptr\n")
			}
		case "CHAR":
			var val = e.basicLit.Value
			var char = val[1]
			if val[1] == '\\' {
				switch val[2] {
				case '\'':
					char = '\''
				case 'n':
					char = '\n'
				case '\\':
					char = '\\'
				case 't':
					char = '\t'
				case 'r':
					char = '\r'
				}
			}
			fmtPrintf("  pushq $%d # convert char literal to int\n", Itoa(int(char)))
		default:
			fmtPrintf("[emitExpr][*astBasicLit] TBI : %s\n", e.basicLit.Kind)
			os.Exit(1)
		}
	case "*astCallExpr":
		fmtPrintf("# [*astCallExpr]\n")
		var fun = e.callExpr.Fun
		if isType(fun) {
			emitConversion(e2t(fun), e.callExpr.Args[0])
			return
		}
		var pushed int = 0
		//var arg *astExpr
		var i int
		for i=len(e.callExpr.Args)-1;i>=0;i-- {
			emitExpr(e.callExpr.Args[i])
			var typ = getTypeOfExpr(e.callExpr.Args[i])
			var size = getSizeOfType(typ)
			pushed = pushed + size
		}
		switch fun.dtype {
		case "*astIdent":
			var ident = fun.ident
			if ident.Name == "print" {
				fmtPrintf("  callq runtime.printstring\n")
				fmtPrintf("  addq $%s, %%rsp # revert \n", Itoa(16))
			} else {
				fmtPrintf("  callq main.%s\n", ident.Name)
				fmtPrintf("  addq $%s, %%rsp # revert \n", Itoa(pushed))
				var obj = ident.Obj
				var decl = obj.Decl
				if decl == nil {
					panic("[emitExpr][*astCallExpr] decl is nil")
				}
				if decl.dtype !=  "*astFuncDecl" {
					panic("[emitExpr][*astCallExpr] decl.dtype is invalid")
				}
				var fndecl = decl.funcDecl
				if fndecl == nil {
					panic("[emitExpr][*astCallExpr] fndecl is nil")
				}
				if fndecl.Sig == nil {
					panic("[emitExpr][*astCallExpr] fndecl.Sig is nil")
				}
				var results = fndecl.Sig.results
				if fndecl.Sig.results == nil {
					fmtPrintf("# [emitExpr] %s sig.results is nil\n", ident.Name)
				} else {
					fmtPrintf("# [emitExpr] %s sig.results.List = %s\n", ident.Name, Itoa(len(fndecl.Sig.results.List)))
				}

				if results != nil && len(results.List) == 1 {
					fmtPrintf("  pushq %%rax # return value\n")
				} else {
					fmtPrintf("   # No results\n")
				}
			}
		case "*astSelectorExpr":
			var selectorExpr = fun.selectorExpr
			if selectorExpr.X.dtype != "*astIdent" {
				fmtPrintf("[emitExpr] TBI selectorExpr.X.dtype=%s\n", selectorExpr.X.dtype)
				os.Exit(1)
			}
			var symbol string = selectorExpr.X.ident.Name + "." + selectorExpr.Sel.Name
			switch symbol {
			case "os.Exit":
				emitCallNonDecl(symbol, e.callExpr.Args)
			case "syscall.Write":
				emitCallNonDecl(symbol, e.callExpr.Args)
			default:
				fmtPrintf("  callq %s.%s\n", selectorExpr.X.ident.Name, selectorExpr.Sel.Name)
				panic("[emitExpr][*astSelectorExpr] Unsupported call to " + symbol)
			}
		default:
			fmtPrintf("[emitExpr] TBI fun.dtype=%s\n", fun.dtype)
			os.Exit(1)
		}
	case "*astUnaryExpr":
		switch e.unaryExpr.Op {
		case "-":
			emitExpr(e.unaryExpr.X)
			fmtPrintf("  popq %%rax # e.X\n")
			fmtPrintf("  imulq $-1, %%rax\n")
			fmtPrintf("  pushq %%rax\n")
		default:
			fmtPrintf("[emitExpr] `TBI:astUnaryExpr:%s\n", e.dtype)
			os.Exit(1)
		}
	case "*astBinaryExpr":
		if kind(getTypeOfExpr(e.binaryExpr.X)) == T_STRING {
			panic("[emitExpr][*astBinaryExpr] TBI T_STRING")
		}
		emitExpr(e.binaryExpr.X) // left
		emitExpr(e.binaryExpr.Y) // right
		switch e.binaryExpr.Op {
		case "+":
			fmtPrintf("  popq %%rcx # right\n")
			fmtPrintf("  popq %%rax # left\n")
			fmtPrintf("  addq %%rcx, %%rax\n")
			fmtPrintf("  pushq %%rax\n")
		case "-":
			fmtPrintf("  popq %%rcx # right\n")
			fmtPrintf("  popq %%rax # left\n")
			fmtPrintf("  subq %%rcx, %%rax\n")
			fmtPrintf("  pushq %%rax\n")
		case "*":
			fmtPrintf("  popq %%rcx # right\n")
			fmtPrintf("  popq %%rax # left\n")
			fmtPrintf("  imulq %%rcx, %%rax\n")
			fmtPrintf("  pushq %%rax\n")
		case "%":
			fmtPrintf("  popq %%rcx # right\n")
			fmtPrintf("  popq %%rax # left\n")
			fmtPrintf("  movq $0, %%rdx # init %%rdx\n")
			fmtPrintf("  divq %%rcx\n")
			fmtPrintf("  movq %%rdx, %%rax\n")
			fmtPrintf("  pushq %%rax\n")
		case "/":
			fmtPrintf("  popq %%rcx # right\n")
			fmtPrintf("  popq %%rax # left\n")
			fmtPrintf("  movq $0, %%rdx # init %%rdx\n")
			fmtPrintf("  divq %%rcx\n")
			fmtPrintf("  pushq %%rax\n")
		default:
			fmtPrintf("# TBI: binary operation for '%s'", e.binaryExpr.Op)
			os.Exit(1)
		}
	default:
		fmtPrintf("[emitExpr] `TBI:%s\n", e.dtype)
		os.Exit(1)
	}
}

func emitListHeadAddr(list *astExpr) {
	var t = getTypeOfExpr(list)
	switch kind(t) {
	case T_ARRAY:
		emitAddr(list) // array head
	case T_SLICE:
		emitExpr(list)
		emitPopSlice()
		fmtPrintf("  pushq %%rax # slice.ptr\n")
	case T_STRING:
		emitExpr(list)
		emitPopString()
		fmtPrintf("  pushq %%rax # string.ptr\n")
	default:
		panic("[emitListHeadAddr] kind=" + kind(getTypeOfExpr(list)))
	}
}

func emitListElementAddr(list *astExpr, elmType *Type) {
	emitListHeadAddr(list)
	emitPopAddress("list head")
	fmtPrintf("  popq %%rcx # index id\n")
	fmtPrintf("  movq $%s, %%rdx # elm size\n", Itoa(getSizeOfType(elmType)))
	fmtPrintf("  imulq %%rdx, %%rcx\n")
	fmtPrintf("  addq %%rcx, %%rax\n")
	fmtPrintf("  pushq %%rax # addr of element\n")
}


func emitVariableAddr(variable *Variable) {
	fmtPrintf("  # variable %s\n", variable.name)

	var addr string
	if variable.isGlobal {
		addr = fmtSprintf("%s(%%rip)", []string{variable.globalSymbol})
	} else {
		addr = fmtSprintf("%d(%%rbp)", []string{Itoa(variable.localOffset)})
	}

	fmtPrintf("  leaq %s, %%rax # addr\n", addr)
	fmtPrintf("  pushq %%rax\n")
}

func emitAddr(expr *astExpr) {
	fmtPrintf("  # emitAddr %s\n", expr.dtype)
	switch expr.dtype {
	case "*astIdent":
		fmtPrintf("  # %s\n", expr.ident.Name)
		if expr.ident.Obj.Kind == "Var" {
			fmtPrintf("  # is Var\n")
			if expr.ident.Obj.Variable == nil {
				fmtPrintf("ERROR: Variable is nil for name %s\n", expr.ident.Obj.Name)
				os.Exit(1)
			}
			emitVariableAddr(expr.ident.Obj.Variable)
		} else {
			fmtPrintf("[emitAddr] Unexpected Kind %s \n" ,expr.ident.Obj.Kind)
		}
	case "*astIndexExpr":
		emitExpr(expr.indexExpr.Index) // index number
		var list = expr.indexExpr.X
		var elmType = getTypeOfExpr(expr)
		emitListElementAddr(list, elmType)
	default:
		panic("[emitAddr] TBI " + expr.dtype)
	}
}

func isType(expr *astExpr) bool {
	switch expr.dtype {
	case "*astArrayType":
		return true
	case "*astIdent":
		if expr.ident == nil {
			panic("[isType] ident should not be nil")
		}
		if expr.ident.Obj == nil {
			panic("[isType] unresolved ident:" + expr.ident.Name)
		}
		fmtPrintf("# [isType][DEBUG] expr.ident.Name = %s\n", expr.ident.Name)
		fmtPrintf("# [isType][DEBUG] expr.ident.Obj = %s,%s\n",
			expr.ident.Obj.Name, expr.ident.Obj.Kind )
		return expr.ident.Obj.Kind == "Typ"
	case "*astParentExpr":
		return true
	}

	return false
}

func emitConversion(tp *Type, arg0 *astExpr) {
	fmtPrintf("# [emitConversion]\n")
	var typeExpr = tp.e
	switch typeExpr.dtype {
	case "*astIdent":
		switch typeExpr.ident.Obj {
		case gInt, gUint8, gUint16, gUintptr: // int(e)
			fmtPrintf("# [emitConversion] to int \n")
			emitExpr(arg0)
		default:
			panic("[emitConversion] TBI : " + typeExpr.ident.Obj.Name)
		}
	case "*astArrayType": // Conversion to slice
		var arrayType = typeExpr.arrayType
		if arrayType.Len != nil {
			panic("[emitConversion] internal error")
		}
		if (kind(getTypeOfExpr(arg0))) != T_STRING {
			panic("source type should be string")
		}
		fmtPrintf("  # Conversion of string => slice \n")
		emitExpr(arg0)
		emitPopString()
		fmtPrintf("  pushq %%rcx # cap\n")
		fmtPrintf("  pushq %%rcx # len\n")
		fmtPrintf("  pushq %%rax # ptr\n")
	default:
		panic("[emitConversion] TBI :" + typeExpr.dtype)
	}
}

func emitLoad(t *Type) {
	if (t == nil) {
		fmtPrintf("# [emitLoad] nil type error\n")
		os.Exit(1)
	}
	emitPopAddress(kind(t))
	switch kind(t) {
	case T_SLICE:
		fmtPrintf("  movq %d(%%rax), %%rdx\n", Itoa(16))
		fmtPrintf("  movq %d(%%rax), %%rcx\n", Itoa(8))
		fmtPrintf("  movq %d(%%rax), %%rax\n", Itoa(0))
		fmtPrintf("  pushq %%rdx # cap\n")
		fmtPrintf("  pushq %%rcx # len\n")
		fmtPrintf("  pushq %%rax # ptr\n")
	case T_STRING:
		fmtPrintf("  movq %d(%%rax), %%rdx\n", Itoa(8))
		fmtPrintf("  movq %d(%%rax), %%rax\n", Itoa(0))
		fmtPrintf("  pushq %%rdx # len\n")
		fmtPrintf("  pushq %%rax # ptr\n")
	case T_UINT8:
		fmtPrintf("  movzbq %d(%%rax), %%rax # load uint8\n", Itoa(0))
		fmtPrintf("  pushq %%rax\n")
	case T_UINT16:
		fmtPrintf("  movzwq %d(%%rax), %%rax # load uint16\n", Itoa(0))
		fmtPrintf("  pushq %%rax\n")
	case T_INT, T_BOOL, T_UINTPTR, T_POINTER:
		fmtPrintf("  movq %d(%%rax), %%rax # load int\n", Itoa(0))
		fmtPrintf("  pushq %%rax\n")

	default:
		fmtPrintf("[emitLoad] TBI:kind=%s\n", kind(t))
		os.Exit(1)
	}
}

func emitStore(t *Type) {
	fmtPrintf("  # emitStore(%s)\n", kind(t))
	switch kind(t) {
	case T_SLICE:
		emitPopSlice()
		fmtPrintf("  popq %%rsi # lhs ptr addr\n")
		fmtPrintf("  movq %%rax, %d(%%rsi) # ptr to ptr\n", Itoa(0))
		fmtPrintf("  movq %%rcx, %d(%%rsi) # len to len\n", Itoa(8))
		fmtPrintf("  movq %%rdx, %d(%%rsi) # cap to cap\n", Itoa(16))
	case T_STRING:
		emitPopString()
		fmtPrintf("  popq %%rsi # lhs ptr addr\n")
		fmtPrintf("  movq %%rax, %d(%%rsi) # ptr to ptr\n", Itoa(0))
		fmtPrintf("  movq %%rcx, %d(%%rsi) # len to len\n", Itoa(8))
	case T_INT:
		fmtPrintf("  popq %%rdi # rhs evaluated\n")
		fmtPrintf("  popq %%rax # lhs addr\n")
		fmtPrintf("  movq %%rdi, (%%rax) # assign\n")
	case T_UINT8:
		fmtPrintf("  popq %%rdi # rhs evaluated\n")
		fmtPrintf("  popq %%rax # lhs addr\n")
		fmtPrintf("  movb %%dil, (%%rax) # assign byte\n")
	default:
		panic("[emitStore] TBI:" + kind(t))
	}
}

func emitAssign(lhs *astExpr, rhs *astExpr) {
	fmtPrintf("  # Assignment: emitAddr(lhs:%s)\n", lhs.dtype)
	emitAddr(lhs)
	fmtPrintf("  # Assignment: emitExpr(rhs)\n")
	emitExpr(rhs)
	emitStore(getTypeOfExpr(lhs))
}


func emitStmt(stmt *astStmt) {
	fmtPrintf("# [emitStmt] dtype=%s\n", stmt.dtype)
	switch stmt.dtype {
	case "*astBlockStmt":
		var stmt2 *astStmt
		for _, stmt2 = range stmt.blockStmt.List {
			emitStmt(stmt2)
		}
	case "*astExprStmt":
		emitExpr(stmt.exprStmt.X)
	case "*astDeclStmt":
		var decl *astDecl = stmt.DeclStmt.Decl
		if decl.dtype != "*astGenDecl" {
			panic("[emitStmt][*astDeclStmt] internal error")
		}
		var genDecl = decl.genDecl
		var valSpec = genDecl.Spec
		var t = e2t(valSpec.Type)
		var ident = valSpec.Name
		var lhs = new(astExpr)
		lhs.dtype = "*astIdent"
		lhs.ident = ident
		var rhs *astExpr
		if valSpec.Value == nil {
			fmtPrintf("  # lhs addresss\n")
			emitAddr(lhs)
			fmtPrintf("  # emitZeroValue for %s\n", t.e.dtype)
			emitZeroValue(t)
			fmtPrintf("  # Assignment: zero value\n")
			emitStore(t)
		} else {
			rhs = valSpec.Value
			emitAssign(lhs, rhs)
		}
		/*
		var valueSpec *astValueSpec = genDecl.Specs[0]
		var obj *astObject = valueSpec.Name.Obj
		var typ *astExpr = valueSpec.Type

		fmtPrintf("[emitStmt] TBI declSpec:%s\n", valueSpec.Name.Name)
		os.Exit(1)

		 */
	case "*astAssignStmt":
		switch stmt.assignStmt.Tok {
		case "=":
			var lhs = stmt.assignStmt.Lhs
			var rhs = stmt.assignStmt.Rhs
			emitAssign(lhs, rhs)
		default:
			fmtPrintf("TBI: assignment of " + stmt.assignStmt.Tok)
			os.Exit(1)
		}
	case "*astReturnStmt":
		if len(stmt.returnStmt.Results) == 0 {
			fmtPrintf("  leave\n")
			fmtPrintf("  ret\n")
		} else if len(stmt.returnStmt.Results) == 1 {
			emitExpr(stmt.returnStmt.Results[0])
			var knd = kind(getTypeOfExpr(stmt.returnStmt.Results[0]))
			switch knd {
			case T_INT:
				fmtPrintf("  popq %%rax # return 64bit\n")
			default:
				panic("[emitStmt][*astReturnStmt] TBI:" + knd)
			}
		} else {
			panic("[emitStmt][*astReturnStmt] TBI\n")
		}
	default:
		fmtPrintf("[emitStmt] TBI:%s\n", stmt.dtype)
		os.Exit(1)
	}
}

func emitFuncDecl(pkgPrefix string, fnc *Func) {
	var localarea = fnc.localarea
	fmtPrintf("\n")
	var fname = fnc.name
	fmtPrintf(pkgPrefix + "." + fname + ":\n")

	fmtPrintf("  pushq %%rbp\n")
	fmtPrintf("  movq %%rsp, %%rbp\n")
	if localarea != 0 {
		fmtPrintf("  subq $%d, %%rsp # local area\n", Itoa(-localarea))
	}

	fmtPrintf("  # func body\n")
	var stmt = new(astStmt)
	stmt.dtype = "*astBlockStmt"
	stmt.blockStmt = fnc.Body
	emitStmt(stmt)

	fmtPrintf("  leave\n")
	fmtPrintf("  ret\n")
}

func emitText(pkgName string) {
	fmtPrintf(".text\n")
	var i int
	for i = 0; i < len(globalFuncs); i++ {
		var fnc = globalFuncs[i]
		emitFuncDecl(pkgName, fnc)
	}
}

type Type struct {
	//kind string
	e *astExpr
}

var tInt *Type
var tUint8 *Type
var tUint16 *Type
var tUintptr *Type
var tString *Type
var tBool *Type

func getTypeOfExpr(expr *astExpr) *Type {
	switch expr.dtype {
	case "*astIdent":
		switch expr.ident.Obj.Kind {
		case "Var":
			switch expr.ident.Obj.Decl.dtype {
			case "*astValueSpec":
				var decl = expr.ident.Obj.Decl.valueSpec
				var t = new(Type)
				t.e = decl.Type
				return t
			case "*astField":
				var decl = expr.ident.Obj.Decl.field
				var t = new(Type)
				t.e = decl.Type
				return t
			default:
				fmtPrintf("[getTypeOfExpr] ERROR 0\n")
				os.Exit(1)
			}
		default:
			fmtPrintf("[getTypeOfExpr] ERROR 1\n")
			os.Exit(1)
		}
	case "*astBasicLit":
		switch expr.basicLit.Kind {
		case "STRING":
			return tString
		case "INT":
			return tInt
		case "CHAR":
			return tInt
		default:
			panic("[getTypeOfExpr] TBI:" + expr.basicLit.Kind)
		}
	case "*astIndexExpr":
		var list = expr.indexExpr.X
		return getElementTypeOfListType(getTypeOfExpr(list))
	case "*astUnaryExpr":
		switch expr.unaryExpr.Op {
		case "-":
			return getTypeOfExpr(expr.unaryExpr.X)
		default:
			panic("[getTypeOfExpr] TBI: Op=" + expr.unaryExpr.Op)
		}
	case "*astCallExpr":
		var fun = expr.callExpr.Fun
		switch fun.dtype {
		case "*astIdent":
			var fn = fun.ident
			if fn.Obj == nil {
				panic("[getTypeOfExpr][astCallExpr] nil Obj is not allowed")
			}
			switch fn.Obj.Kind {
			case "Typ":
				return e2t(fun)
			case "Fun":
				var decl = fn.Obj.Decl
				switch decl.dtype {
				case "*astFuncDecl":
					var resultList = decl.funcDecl.Sig.results.List
					if len(resultList) != 1 {
						panic("[getTypeOfExpr][astCallExpr] len results.List is not 1")
					}
					return e2t(decl.funcDecl.Sig.results.List[0].Type)
				default:
					panic("[getTypeOfExpr][astCallExpr] decl.dtype=" + decl.dtype)
				}
				panic("[getTypeOfExpr][astCallExpr] Fun ident " + fn.Name)
			}
		default:
			panic("[getTypeOfExpr][astCallExpr] dtype=" + expr.callExpr.Fun.dtype)
		}
	case "*astBinaryExpr":
		switch expr.binaryExpr.Op {
		case "==", "!=", "<", ">", "<=", ">=":
			return tBool
		default:
			return getTypeOfExpr(expr.binaryExpr.X)
		}
	default:
		fmtPrintf("[getTypeOfExpr] dtype=%s\n", expr.dtype)
		os.Exit(1)
	}

	fmtPrintf("[getTypeOfExpr] nil type is not allowed\n")
	os.Exit(1)
	var r *Type
	return r
}

func generateCode(pkgName string) {
	fmtPrintf("")
	fmtPrintf("runtime.heapInit:\n")
	fmtPrintf("ret\n")

	emitData(pkgName)
	emitText(pkgName)
}

type Func struct {
	localvars []*string
	localarea int
	argsarea  int
	name      string
	Body *astBlockStmt
}

type sliteral struct {
	label  string
	strlen int
	value  string // raw value
}

var globalVars []*astValueSpec
var globalFuncs []*Func

type astFile struct {
	Name string
	Decls []*astDecl
	Unresolved []*astIdent
}

func kind(t *Type) string {
	if t == nil {
		fmtPrintf("nil type is not expected\n")
		os.Exit(1)
	}
	var e = t.e
	switch t.e.dtype {
	case "*astIdent":
		var ident = t.e.ident
		switch ident.Name {
		case "uintptr":
			return T_UINTPTR
		case "int":
			return T_INT
		case "string":
			return T_STRING
		case "uint8":
			return T_UINT8
		case "uint16":
			return T_UINT16
		case "bool":
			return T_BOOL
		default:
			fmtPrintf("[kind] unsupported type %s\n", ident.Name)
			os.Exit(1)
		}
	case "*astArrayType":
		if e.arrayType.Len == nil {
			return T_SLICE
		} else {
			return T_ARRAY
		}
	default:
		panic("[kind] Unkown dtype:" + t.e.dtype)
	}
	fmtPrintf("error")
	os.Exit(1)
	return ""
}

func main() {
	initGlobals()
	var sourceFiles = []string{"2gen/sample.go"}
	var sourceFile string
	for _, sourceFile = range sourceFiles {
		globalVars = nil
		globalFuncs = nil
		stringLiterals = nil
		stringIndex = 0
		var f = parseFile(sourceFile)
		var pkgName = semanticAnalyze(f)
		//panic("[main] STOP for debug")
		generateCode(pkgName)
	}
}

func initGlobals() {
	T_STRING  = "T_STRING"
	T_SLICE  = "T_SLICE"
	T_BOOL  = "T_BOOL"
	T_INT  = "T_INT"
	T_UINT8  = "T_UINT8"
	T_UINT16  = "T_UINT16"
	T_UINTPTR  = "T_UINTPTR"
	T_ARRAY  = "T_ARRAY"
	T_STRUCT  = "T_STRUCT"
	T_POINTER  = "T_POINTER"

	gString = new(astObject)
	gString.Kind = "Typ"
	gString.Name = "string"
	tString = new(Type)
	tString.e = new(astExpr)
	tString.e.dtype = "*astIdent"
	tString.e.ident = new(astIdent)
	tString.e.ident.Name = "string"
	tString.e.ident.Obj = gString

	gInt = new(astObject)
	gInt.Kind = "Typ"
	gInt.Name = "int"
	tInt = new(Type)
	tInt.e = new(astExpr)
	tInt.e.dtype = "*astIdent"
	tInt.e.ident = new(astIdent)
	tInt.e.ident.Name = "int"
	tInt.e.ident.Obj = gInt

	gUint8 = new(astObject)
	gUint8.Kind = "Typ"
	gUint8.Name = "uint8"
	tUint8 = new(Type)
	tUint8.e = new(astExpr)
	tUint8.e.dtype = "*astIdent"
	tUint8.e.ident = new(astIdent)
	tUint8.e.ident.Name = "uint8"
	tUint8.e.ident.Obj = gUint8

	gUint16 = new(astObject)
	gUint16.Kind = "Typ"
	gUint16.Name = "uint16"
	tUint16 = new(Type)
	tUint16.e = new(astExpr)
	tUint16.e.dtype = "*astIdent"
	tUint16.e.ident = new(astIdent)
	tUint16.e.ident.Name = "uint16"
	tUint16.e.ident.Obj = gUint16

	gUintptr = new(astObject)
	gUintptr.Kind = "Typ"
	gUintptr.Name = "uintptr"
	tUintptr = new(Type)
	tUintptr.e = new(astExpr)
	tUintptr.e.dtype = "*astIdent"
	tUintptr.e.ident = new(astIdent)
	tUintptr.e.ident.Name = "uintptr"
	tUintptr.e.ident.Obj = gUintptr

	gBool = new(astObject)
	gBool.Kind = "Typ"
	gBool.Name = "bool"
	tBool = new(Type)
	tBool.e = new(astExpr)
	tBool.e.dtype = "*astIdent"
	tBool.e.ident = new(astIdent)
	tBool.e.ident.Name = "bool"
	tBool.e.ident.Obj = gBool
}
