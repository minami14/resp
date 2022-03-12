package resp

import (
	"bufio"
	"bytes"
	"strconv"
)

type ResponseWriter struct {
	writer *bufio.Writer
}

func SerializeResponse(resp Response) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	w := bufio.NewWriter(buf)
	if err := resp.WriteResponse(w); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (r *ResponseWriter) WriteResponse(resp Response) error {
	if resp == nil {
		resp = new(NullResponse)
	}
	return resp.WriteResponse(r.writer)
}

func (r *ResponseWriter) WriteString(v string) error {
	resp := &StringResponse{Value: v}
	return resp.WriteResponse(r.writer)
}

func (r *ResponseWriter) WriteInt(v int) error {
	resp := &IntResponse{Value: v}
	return resp.WriteResponse(r.writer)
}

func (r *ResponseWriter) WriteBulkString(v string) error {
	resp := &BulkStringResponse{Value: v}
	return resp.WriteResponse(r.writer)
}

func (r *ResponseWriter) WriteNull() error {
	resp := new(NullResponse)
	return resp.WriteResponse(r.writer)
}

func (r *ResponseWriter) WriteError(v string) error {
	resp := &ErrorResponse{Value: v}
	return resp.WriteResponse(r.writer)
}

type Response interface {
	WriteResponse(*bufio.Writer) error
}

func writeNewLine(w *bufio.Writer) error {
	if _, err := w.WriteString("\r\n"); err != nil {
		return err
	}
	return nil
}

type StringResponse struct {
	Value string
}

func (r *StringResponse) WriteResponse(w *bufio.Writer) error {
	if err := w.WriteByte('+'); err != nil {
		return err
	}
	if _, err := w.WriteString(r.Value); err != nil {
		return err
	}
	if err := writeNewLine(w); err != nil {
		return err
	}
	return w.Flush()
}

type ErrorResponse struct {
	Value string
}

func (r *ErrorResponse) WriteResponse(w *bufio.Writer) error {
	if err := w.WriteByte('-'); err != nil {
		return err
	}
	if _, err := w.WriteString(r.Value); err != nil {
		return err
	}
	if err := writeNewLine(w); err != nil {
		return err
	}
	return w.Flush()
}

type IntResponse struct {
	Value int
}

func (r *IntResponse) WriteResponse(w *bufio.Writer) error {
	if err := w.WriteByte(':'); err != nil {
		return err
	}
	if _, err := w.WriteString(strconv.Itoa(r.Value)); err != nil {
		return err
	}
	if err := writeNewLine(w); err != nil {
		return err
	}
	return w.Flush()
}

type BulkStringResponse struct {
	Value string
}

func (r *BulkStringResponse) WriteResponse(w *bufio.Writer) error {
	if err := w.WriteByte('$'); err != nil {
		return err
	}
	if _, err := w.WriteString(strconv.Itoa(len(r.Value))); err != nil {
		return err
	}
	if err := writeNewLine(w); err != nil {
		return err
	}
	if _, err := w.WriteString(r.Value); err != nil {
		return err
	}
	if err := writeNewLine(w); err != nil {
		return err
	}
	return w.Flush()
}

type NullResponse struct{}

func (r *NullResponse) WriteResponse(w *bufio.Writer) error {
	if _, err := w.WriteString("$-1\r\n"); err != nil {
		return err
	}
	return w.Flush()
}

type ArrayResponse struct {
	Responses []Response
}

func (r *ArrayResponse) WriteResponse(w *bufio.Writer) error {
	if r.Responses == nil {
		if _, err := w.WriteString("*-1\\r\\n"); err != nil {
			return err
		}
		return nil
	}

	if err := w.WriteByte('*'); err != nil {
		return err
	}
	if _, err := w.WriteString(strconv.Itoa(len(r.Responses))); err != nil {
		return err
	}
	if err := writeNewLine(w); err != nil {
		return err
	}
	for _, resp := range r.Responses {
		if err := resp.WriteResponse(w); err != nil {
			return err
		}
	}
	return w.Flush()
}
