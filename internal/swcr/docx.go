package swcr

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

// ---------------------------------------------------------------------------
// OOXML namespace constants
// ---------------------------------------------------------------------------

const (
	nsW       = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	nsR       = "http://schemas.openxmlformats.org/package/2006/relationships"
	nsContent = "http://schemas.openxmlformats.org/package/2006/content-types"
	nsRel     = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
	nsCore    = "http://schemas.openxmlformats.org/package/2006/metadata/core-properties"
	nsDC      = "http://purl.org/dc/elements/1.1/"
	nsDCTerms = "http://purl.org/dc/terms/"
	nsXSI     = "http://www.w3.org/2001/XMLSchema-instance"
)

// ---------------------------------------------------------------------------
// [Content_Types].xml
// ---------------------------------------------------------------------------

type ctOverride struct {
	XMLName     struct{} `xml:"Override"`
	PartName    string   `xml:"PartName,attr"`
	ContentType string   `xml:"ContentType,attr"`
}

type ctDefault struct {
	XMLName     struct{} `xml:"Default"`
	Extension   string   `xml:"Extension,attr"`
	ContentType string   `xml:"ContentType,attr"`
}

type ctTypes struct {
	XMLName   struct{}     `xml:"http://schemas.openxmlformats.org/package/2006/content-types Types"`
	Defaults  []ctDefault  `xml:"Default"`
	Overrides []ctOverride `xml:"Override"`
}

// ---------------------------------------------------------------------------
// _rels/.rels
// ---------------------------------------------------------------------------

type relEntry struct {
	XMLName struct{} `xml:"Relationship"`
	ID      string   `xml:"Id,attr"`
	Type    string   `xml:"Type,attr"`
	Target  string   `xml:"Target,attr"`
}

type relRels struct {
	XMLName struct{}   `xml:"Relationships"`
	Xmlns   string     `xml:"xmlns,attr"`
	Entries []relEntry `xml:"Relationship"`
}

// ---------------------------------------------------------------------------
// word/_rels/document.xml.rels
// ---------------------------------------------------------------------------

type docRels struct {
	XMLName struct{}   `xml:"http://schemas.openxmlformats.org/package/2006/relationships Relationships"`
	Entries []relEntry `xml:"Relationship"`
}

// ---------------------------------------------------------------------------
// word/document.xml — wordprocessingml elements
//
// OOXML requires the root to be <w:document> containing <w:body>, with
// <w:sectPr> as the LAST child of <w:body>.  To avoid namespace bloat, we
// declare xmlns:w on the root element and use the explicit w: prefix in
// every child tag.  xml.Encoder (encodeWordXML) emits the prefix correctly
// without duplicating the namespace on every element.
// ---------------------------------------------------------------------------

// wDocument is the root element of word/document.xml.
type wDocument struct {
	XMLName struct{} `xml:"w:document"`
	W       string   `xml:"xmlns:w,attr"`
	R       string   `xml:"xmlns:r,attr"`
	Body    wBody    `xml:"w:body"`
}

type wBody struct {
	XMLName    struct{} `xml:"w:body"`
	Paragraphs []wP     `xml:"w:p"`      // before SectPr — ordering matters
	SectPr     *wSectPr `xml:"w:sectPr"` // MUST be the last child
}

type wSectPr struct {
	XMLName     struct{}   `xml:"w:sectPr"`
	HeaderRefs  []wHdrRef  `xml:"w:headerReference"`
	FooterRefs  []wFtrRef  `xml:"w:footerReference,omitempty"`
	PageSize    *wPgSz     `xml:"w:pgSz,omitempty"`
	PageMargins *wPgMar    `xml:"w:pgMar,omitempty"`
	Type        *wSectType `xml:"w:type,omitempty"`
}

type wHdrRef struct {
	XMLName struct{} `xml:"w:headerReference"`
	Type    string   `xml:"w:type,attr"`
	ID      string   `xml:"r:id,attr"`
}

type wFtrRef struct {
	XMLName struct{} `xml:"w:footerReference"`
	Type    string   `xml:"w:type,attr"`
	ID      string   `xml:"r:id,attr"`
}

type wPgSz struct {
	XMLName struct{} `xml:"w:pgSz"`
	Width   int      `xml:"w:w,attr"`
	Height  int      `xml:"w:h,attr"`
	Orient  string   `xml:"w:orient,attr,omitempty"`
}

type wPgMar struct {
	XMLName struct{} `xml:"w:pgMar"`
	Top     int      `xml:"w:top,attr"`
	Right   int      `xml:"w:right,attr"`
	Bottom  int      `xml:"w:bottom,attr"`
	Left    int      `xml:"w:left,attr"`
	Header  int      `xml:"w:header,attr"`
	Footer  int      `xml:"w:footer,attr"`
}

type wSectType struct {
	XMLName struct{} `xml:"w:type"`
	Val     string   `xml:"w:val,attr"`
}

// ---------------------------------------------------------------------------
// Paragraph
// ---------------------------------------------------------------------------

type wP struct {
	XMLName    struct{} `xml:"w:p"`
	Properties *wPPr    `xml:"w:pPr,omitempty"`
	Runs       []wR     `xml:"w:r,omitempty"`
}

type wPPr struct {
	XMLName struct{}  `xml:"w:pPr"`
	Spacing *wSpacing `xml:"w:spacing,omitempty"`
	Jc      *wJc      `xml:"w:jc,omitempty"`
	Tabs    *wTabs    `xml:"w:tabs,omitempty"`
}

type wSpacing struct {
	XMLName  struct{} `xml:"w:spacing"`
	Before   int      `xml:"w:before,attr,omitempty"`
	After    int      `xml:"w:after,attr,omitempty"`
	Line     int      `xml:"w:line,attr,omitempty"`
	LineRule string   `xml:"w:lineRule,attr,omitempty"`
}

type wJc struct {
	XMLName struct{} `xml:"w:jc"`
	Val     string   `xml:"w:val,attr"`
}

type wTabs struct {
	XMLName struct{} `xml:"w:tabs"`
	Tabs    []wTab   `xml:"w:tab"`
}

type wTab struct {
	XMLName struct{} `xml:"w:tab"`
	Val     string   `xml:"w:val,attr"`
	Pos     int      `xml:"w:pos,attr"`
}

// ---------------------------------------------------------------------------
// Run (text span)
// ---------------------------------------------------------------------------

type wR struct {
	XMLName    struct{}    `xml:"w:r"`
	Properties *wRPr       `xml:"w:rPr,omitempty"`
	Text       *wT         `xml:"w:t,omitempty"`
	Tab        *wTabElem   `xml:"w:tab,omitempty"`
	FldChar    *wFldChar   `xml:"w:fldChar,omitempty"`
	InstrText  *wInstrText `xml:"w:instrText,omitempty"`
}

type wRPr struct {
	XMLName struct{} `xml:"w:rPr"`
	RFonts  *wRFonts `xml:"w:rFonts,omitempty"`
	Sz      *wSz     `xml:"w:sz,omitempty"`
	SzCs    *wSzCs   `xml:"w:szCs,omitempty"`
}

type wRFonts struct {
	XMLName  struct{} `xml:"w:rFonts"`
	ASCII    string   `xml:"w:ascii,attr,omitempty"`
	HAnsi    string   `xml:"w:hAnsi,attr,omitempty"`
	EastAsia string   `xml:"w:eastAsia,attr,omitempty"`
	CS       string   `xml:"w:cs,attr,omitempty"`
}

type wSz struct {
	XMLName struct{} `xml:"w:sz"`
	Val     int      `xml:"w:val,attr"`
}

type wSzCs struct {
	XMLName struct{} `xml:"w:szCs"`
	Val     int      `xml:"w:val,attr"`
}

type wT struct {
	XMLName struct{} `xml:"w:t"`
	Space   string   `xml:"xml:space,attr,omitempty"`
	Value   string   `xml:",chardata"`
}

type wTabElem struct {
	XMLName struct{} `xml:"w:tab"`
}

type wFldChar struct {
	XMLName     struct{} `xml:"w:fldChar"`
	FldCharType string   `xml:"w:fldCharType,attr"`
}

type wInstrText struct {
	XMLName struct{} `xml:"w:instrText"`
	Space   string   `xml:"xml:space,attr,omitempty"`
	Value   string   `xml:",chardata"`
}

// ---------------------------------------------------------------------------
// Header (word/header1.xml)
// ---------------------------------------------------------------------------

type wHdr struct {
	XMLName    struct{} `xml:"w:hdr"`
	W          string   `xml:"xmlns:w,attr"`
	Paragraphs []wP     `xml:"w:p"`
}

// ---------------------------------------------------------------------------
// docProps/core.xml
// ---------------------------------------------------------------------------

type coreProperties struct {
	XMLName struct{} `xml:"http://schemas.openxmlformats.org/package/2006/metadata/core-properties coreProperties"`
	Dc      string   `xml:"xmlns:dc,attr"`
	Dcterms string   `xml:"xmlns:dcterms,attr"`
	XSI     string   `xml:"xmlns:xsi,attr"`
}

// ---------------------------------------------------------------------------
// docProps/app.xml
// ---------------------------------------------------------------------------

type appProperties struct {
	XMLName struct{} `xml:"http://schemas.openxmlformats.org/officeDocument/2006/extended-properties Properties"`
}

// ---------------------------------------------------------------------------
// Document — the top-level builder
// ---------------------------------------------------------------------------

// Document is a builder for a .docx file. Call AddParagraph to add a code
// line, SetHeader to set the page header, then Save to write the .docx.
type Document struct {
	wr          *zip.Writer
	paragraphs  []wP
	headerParas []wP
	opts        *WriterOpts
	title       string
}

// WriterOpts controls formatting of generated paragraphs.
type WriterOpts struct {
	FontName    string  // e.g. "宋体" (SimSun)
	FontSize    float64 // in points, e.g. 10.5
	SpaceBefore float64 // paragraph space before, in pts
	SpaceAfter  float64 // paragraph space after, in pts
	LineSpacing float64 // fixed line spacing, in pts
}

// DefaultWriterOpts returns the tuned settings that produce exactly 50 lines
// per A4 page (tested with 宋体 10.5pt).
func DefaultWriterOpts() *WriterOpts {
	return &WriterOpts{
		FontName:    "宋体",
		FontSize:    10.5,
		SpaceBefore: 0,
		SpaceAfter:  2.3,
		LineSpacing: 10.5,
	}
}

// ptToHalfPt converts points to half-points (used for font size in OOXML).
func ptToHalfPt(pt float64) int {
	return int(pt * 2)
}

// ptToTwips converts points to twips (1pt = 20 twips).
func ptToTwips(pt float64) int {
	return int(pt * 20)
}

// NewDocument creates a new .docx builder writing to the given file path.
func NewDocument(path string, opts *WriterOpts) (*Document, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create docx file: %w", err)
	}
	zw := zip.NewWriter(f)
	return &Document{
		wr:   zw,
		opts: opts,
	}, nil
}

// SetHeader sets the page header text (title, centered) and page number
// (right-aligned).
func (d *Document) SetHeader(title string) {
	d.title = title
	sz := ptToHalfPt(d.opts.FontSize)
	fontName := d.opts.FontName

	rf := &wRFonts{ASCII: fontName, HAnsi: fontName, EastAsia: fontName, CS: fontName}
	rp := &wRPr{RFonts: rf, Sz: &wSz{Val: sz}, SzCs: &wSzCs{Val: sz}}

	// Tabs: center at half page width, right at page width.
	// A4 width: 11906 twips. Default margins: 1440 twips each side.
	// Usable width: 11906 - 2*1440 = 9026 twips.
	pageWidth := 11906
	margin := 1440
	bodyWidth := pageWidth - 2*margin

	tabs := &wTabs{Tabs: []wTab{
		{Val: "center", Pos: bodyWidth / 2},
		{Val: "right", Pos: bodyWidth},
	}}

	headerPara := wP{
		Properties: &wPPr{
			Tabs: tabs,
		},
		Runs: []wR{
			// Tab to center
			{Properties: rp, Tab: &wTabElem{}},
			// Title text
			{Properties: rp, Text: &wT{Space: "preserve", Value: title}},
			// Tab to right
			{Properties: rp, Tab: &wTabElem{}},
			// PAGE field: begin
			{Properties: rp, FldChar: &wFldChar{FldCharType: "begin"}},
			// PAGE field: instruction
			{Properties: rp, InstrText: &wInstrText{Space: "preserve", Value: " PAGE "}},
			// PAGE field: separate
			{Properties: rp, FldChar: &wFldChar{FldCharType: "separate"}},
			// PAGE field: display value (1 as placeholder)
			{Properties: rp, Text: &wT{Value: "1"}},
			// PAGE field: end
			{Properties: rp, FldChar: &wFldChar{FldCharType: "end"}},
		},
	}

	d.headerParas = []wP{headerPara}
}

// AddParagraph adds a single code line as a paragraph.
func (d *Document) AddParagraph(text string) {
	sz := ptToHalfPt(d.opts.FontSize)
	fontName := d.opts.FontName

	rf := &wRFonts{ASCII: fontName, HAnsi: fontName, EastAsia: fontName, CS: fontName}
	rp := &wRPr{RFonts: rf, Sz: &wSz{Val: sz}, SzCs: &wSzCs{Val: sz}}

	spacing := &wSpacing{
		Before:   ptToTwips(d.opts.SpaceBefore),
		After:    ptToTwips(d.opts.SpaceAfter),
		Line:     ptToTwips(d.opts.LineSpacing),
		LineRule: "exact",
	}

	p := wP{
		Properties: &wPPr{Spacing: spacing},
		Runs: []wR{
			{Properties: rp, Text: &wT{Space: "preserve", Value: text}},
		},
	}
	d.paragraphs = append(d.paragraphs, p)
}

// Save writes all OOXML parts into the ZIP and closes the file.
func (d *Document) Save() error {
	if err := d.writeContentTypes(); err != nil {
		return err
	}
	if err := d.writeRootRels(); err != nil {
		return err
	}
	if err := d.writeDocRels(); err != nil {
		return err
	}
	if err := d.writeDocument(); err != nil {
		return err
	}
	if err := d.writeHeader(); err != nil {
		return err
	}
	if err := d.writeCoreProps(); err != nil {
		return err
	}
	if err := d.writeAppProps(); err != nil {
		return err
	}
	return d.wr.Close()
}

func (d *Document) writeContentTypes() error {
	w, err := d.wr.Create("[Content_Types].xml")
	if err != nil {
		return err
	}
	types := ctTypes{
		Defaults: []ctDefault{
			{Extension: "rels", ContentType: "application/vnd.openxmlformats-package.relationships+xml"},
			{Extension: "xml", ContentType: "application/xml"},
		},
		Overrides: []ctOverride{
			{PartName: "/word/document.xml", ContentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"},
			{PartName: "/word/header1.xml", ContentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"},
			{PartName: "/docProps/core.xml", ContentType: "application/vnd.openxmlformats-package.core-properties+xml"},
			{PartName: "/docProps/app.xml", ContentType: "application/vnd.openxmlformats-officedocument.extended-properties+xml"},
		},
	}
	return marshalEncode(w, types)
}

func (d *Document) writeRootRels() error {
	w, err := d.wr.Create("_rels/.rels")
	if err != nil {
		return err
	}
	rels := relRels{
		Xmlns: nsR,
		Entries: []relEntry{
			{ID: "rId1", Type: nsRel + "/officeDocument", Target: "word/document.xml"},
			{ID: "rId2", Type: nsRel + "/extended-properties", Target: "docProps/app.xml"},
			{ID: "rId3", Type: nsRel + "/core-properties", Target: "docProps/core.xml"},
		},
	}
	return marshalEncode(w, rels)
}

func (d *Document) writeDocRels() error {
	w, err := d.wr.Create("word/_rels/document.xml.rels")
	if err != nil {
		return err
	}
	rels := docRels{
		Entries: []relEntry{
			{ID: "rId1", Type: nsRel + "/header", Target: "header1.xml"},
		},
	}
	return marshalEncode(w, rels)
}

func (d *Document) writeDocument() error {
	w, err := d.wr.Create("word/document.xml")
	if err != nil {
		return err
	}

	pageWidth := 11906
	pageHeight := 16838
	margin := 1440

	sectPr := &wSectPr{
		HeaderRefs: []wHdrRef{{Type: "default", ID: "rId1"}},
		PageSize:   &wPgSz{Width: pageWidth, Height: pageHeight},
		PageMargins: &wPgMar{
			Top:    margin,
			Right:  margin,
			Bottom: margin,
			Left:   margin,
			Header: 851,
			Footer: 851,
		},
	}

	doc := wDocument{
		W: nsW,
		R: nsRel,
		Body: wBody{
			Paragraphs: d.paragraphs,
			SectPr:     sectPr,
		},
	}
	return encodeWordXML(w, doc)
}

func (d *Document) writeHeader() error {
	w, err := d.wr.Create("word/header1.xml")
	if err != nil {
		return err
	}

	hdr := wHdr{
		W:          nsW,
		Paragraphs: d.headerParas,
	}
	return encodeWordXML(w, hdr)
}

func (d *Document) writeCoreProps() error {
	w, err := d.wr.Create("docProps/core.xml")
	if err != nil {
		return err
	}
	props := coreProperties{
		Dc:      nsDC,
		Dcterms: nsDCTerms,
		XSI:     nsXSI,
	}
	return marshalEncode(w, props)
}

func (d *Document) writeAppProps() error {
	w, err := d.wr.Create("docProps/app.xml")
	if err != nil {
		return err
	}
	return marshalEncode(w, appProperties{})
}

// ---------------------------------------------------------------------------
// XML encoding helpers
// ---------------------------------------------------------------------------

// marshalEncode writes v using xml.MarshalIndent prefixed by the XML header.
// Used for non-wordprocessingml parts where default namespace is acceptable.
func marshalEncode(w io.Writer, v any) error {
	data, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if _, writeErr := w.Write([]byte(xml.Header)); writeErr != nil {
		return writeErr
	}
	_, err = w.Write(data)
	return err
}

// encodeWordXML writes v using xml.Encoder, which is namespace-aware.
// The root element must declare xmlns:w via a struct field tagged
// `xml:"xmlns:w,attr"`. All child wordprocessingml elements use `w:`
// prefix in their struct tags; the encoder emits the prefix without
// duplicating the namespace declaration.
func encodeWordXML(w io.Writer, v any) error {
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return err
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	return enc.Encode(v)
}
