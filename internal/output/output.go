package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

type ContextSummary struct {
	Alias     string `json:"alias,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
}

type ErrorPayload struct {
	Type       string `json:"type,omitempty"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code,omitempty"`
	Retryable  bool   `json:"retryable,omitempty"`
	DocURL     string `json:"doc_url,omitempty"`
	BackoffMS  *int   `json:"backoff_ms,omitempty"`
}

type Meta struct {
	RequestID  string `json:"request_id,omitempty"`
	Pagination any    `json:"pagination,omitempty"`
}

type Envelope struct {
	OK      bool            `json:"ok"`
	Context *ContextSummary `json:"context,omitempty"`
	Data    any             `json:"data"`
	Meta    Meta            `json:"meta"`
	Error   *ErrorPayload   `json:"error"`
}

func Success(ctx *ContextSummary, data any, meta Meta) Envelope {
	return Envelope{
		OK:      true,
		Context: ctx,
		Data:    data,
		Meta:    meta,
		Error:   nil,
	}
}

func Failure(ctx *ContextSummary, meta Meta, err *ErrorPayload) Envelope {
	return Envelope{
		OK:      false,
		Context: ctx,
		Data:    nil,
		Meta:    meta,
		Error:   err,
	}
}

func PrintJSON(w io.Writer, payload any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}

func PrintTable(w io.Writer, rows []map[string]string) error {
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "no rows")
		return err
	}

	headers := orderedHeaders(rows)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, strings.Join(headers, "\t")); err != nil {
		return err
	}
	for _, row := range rows {
		values := make([]string, 0, len(headers))
		for _, header := range headers {
			values = append(values, row[header])
		}
		if _, err := fmt.Fprintln(tw, strings.Join(values, "\t")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func orderedHeaders(rows []map[string]string) []string {
	set := map[string]struct{}{}
	for _, row := range rows {
		for key := range row {
			set[key] = struct{}{}
		}
	}

	headers := make([]string, 0, len(set))
	for key := range set {
		headers = append(headers, key)
	}
	sort.Strings(headers)
	return headers
}

func MustPrintJSON(payload any) {
	if err := PrintJSON(os.Stdout, payload); err != nil {
		fmt.Fprintf(os.Stderr, "print json: %v\n", err)
	}
}
