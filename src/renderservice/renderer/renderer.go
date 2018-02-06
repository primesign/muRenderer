package renderer

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"unsafe"

	log "github.com/Sirupsen/logrus"
)

// #cgo CFLAGS: -I./mupdf/include -I./mupdf/include/mupdf -I./mupdf/thirdparty/openjpeg -I./mupdf/thirdparty/jbig2dec -I./mupdf/thirdparty/zlib -I./mupdf/thirdparty/jpeg -I./mupdf/thirdparty/freetype -I./mupdf/thirdparty/freeglut
// #cgo LDFLAGS: -L${SRCDIR}/mupdf/build/release/ -lmupdf -lmupdfthird -lm
// #include <fitz.h>
import "C"

var (
	// the fitz library version in use
	fzVersion = C.CString("1.12")
	// gives filetype when opening stream
	magic = C.CString("pdf")
)

// Renderer represents the inteface all renderer implementations have to match
type Renderer interface {
	NumPages() int
	PageInfo(pageNr int) *PageInfo
	RenderPageForZoom(pageNr int, zoom float64) (io.Reader, error)
	RenderPageForRectangle(pageNr int, maxWidth int, maxHeight int) (io.Reader, error)
	CloseRenderer()
}

// MuRenderer respresents an instance of a mupdf renderer for a specific document
type MuRenderer struct {
	ctx    *C.fz_context
	cDoc   *C.fz_document
	stream *C.fz_stream
}

// PageInfo holds information about the document
type PageInfo struct {
	Width  float64
	Height float64
}

// NewRenderer returns a pointer to a renderer for the given buffer
func NewRenderer(buf []byte) Renderer {
	log.Debug("Renderer: call new")
	r := new(MuRenderer)
	r.ctx = C.fz_new_context_imp(nil, nil, C.FZ_STORE_UNLIMITED, fzVersion)
	charPtr := (*C.uchar)(unsafe.Pointer(&buf[0]))
	r.stream = C.fz_open_memory(r.ctx, charPtr, C.int(binary.Size(buf)))
	C.fz_register_document_handlers(r.ctx)
	r.cDoc = C.fz_open_document_with_stream(r.ctx, magic, r.stream)
	return Renderer(r)
}

// NumPages returns the number of pages
func (r *MuRenderer) NumPages() int {
	log.Debug("Renderer: get number of pages")
	return int(C.fz_count_pages(r.ctx, r.cDoc))
}

// PageInfo returns the width and height  of the specified page
func (r *MuRenderer) PageInfo(pageNr int) *PageInfo {
	log.Debug("Renderer: get page info")
	page := C.fz_load_page(r.ctx, r.cDoc, C.int(pageNr-1))
	defer C.fz_drop_page(r.ctx, page)
	re := new(C.fz_rect)
	rect := C.fz_bound_page(r.ctx, page, re)

	return &PageInfo{
		Width:  float64(rect.x1 - rect.x0),
		Height: float64(rect.y1 - rect.y0),
	}
}

// RenderPageForZoom renders the specified page of the document with teh given zoom level
func (r *MuRenderer) RenderPageForZoom(pageNr int, zoom float64) (io.Reader, error) {
	log.Debug("Renderer: render page with zoom level")
	return r.renderPng(pageNr, zoom)
}

// RenderPageForRectangle renders the specified page in such a way, that the result does not exceed maxWidth/maxHeight
// the aspect ratio is preserved
func (r *MuRenderer) RenderPageForRectangle(pageNr int, maxWidth int, maxHeight int) (io.Reader, error) {
	log.Debug(fmt.Sprintf("Renderer: render page for width %d height %d", maxWidth, maxHeight))
	if maxWidth == 0 || maxHeight == 0 {
		return nil, errors.New("Neither maxWidth nor maxHeight can be zero.")
	}
	pi := r.PageInfo(pageNr)
	zoomX := float64(maxWidth) / pi.Width
	zoomY := float64(maxHeight) / pi.Height

	return r.renderPng(pageNr, math.Min(zoomX, zoomY))
}

func (r *MuRenderer) renderPng(pageNr int, zoom float64) (io.Reader, error) {

	page := C.fz_load_page(r.ctx, r.cDoc, C.int(pageNr-1))
	defer C.fz_drop_page(r.ctx, page)

	bounds := new(C.fz_rect)
	C.fz_bound_page(r.ctx, page, bounds)

	var transform C.fz_matrix
	C.fz_scale(&transform, C.float(zoom), C.float(zoom))

	C.fz_transform_rect(bounds, &transform)

	bbox := new(C.fz_irect)
	C.fz_round_rect(bbox, bounds)
	pix := C.fz_new_pixmap_with_bbox(r.ctx, C.fz_device_rgb(r.ctx), bbox)
	defer C.fz_drop_pixmap(r.ctx, pix)
	C.fz_clear_pixmap_with_value(r.ctx, pix, C.int(0xff))

	dev := C.fz_new_draw_device(r.ctx, pix)
	defer C.fz_drop_device(r.ctx, dev)

	C.fz_run_page(r.ctx, page, dev, &transform, nil)

	buf := C.fz_new_buffer_from_pixmap_as_png(r.ctx, pix)
	pngBuf := C.GoBytes(unsafe.Pointer(buf.data), buf.len)
	return bytes.NewReader(pngBuf), nil
}

// CloseRenderer frees the renderers resources
func (r *MuRenderer) CloseRenderer() {
	log.Debug("Renderer: close renderer")
	C.fz_drop_document(r.ctx, r.cDoc)
	C.fz_drop_stream(r.ctx, r.stream)
	C.fz_drop_context(r.ctx)
}
