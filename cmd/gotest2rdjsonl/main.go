package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	inputFile string
	bufLines  uint
)

func run() int {
	flag.StringVar(&inputFile, "input", "", "input file (default: stdin)")
	flag.UintVar(&bufLines, "buf-lines", 3, "the number of buffering lines")
	flag.Parse()

	input, err := openInputStream(inputFile)
	if err != nil {
		slog.Error("failed to open input stream", slog.String("error", err.Error()))
		return 1
	}
	defer input.Close()
	dec := json.NewDecoder(input)
	enc := json.NewEncoder(os.Stdout)
	var diags []RDFDiagnostic
	bufs := make([]TestEvent, 0, bufLines+1)
	for {
		var ev TestEvent
		err := dec.Decode(&ev)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			slog.Error("failed to parse JSON line", slog.String("error", err.Error()))
			return 1
		}
		slog.Debug("parsed event", slog.String("action", ev.Action), slog.String("test", ev.Test), slog.String("package", ev.Package), slog.String("output", ev.Output))
		switch ev.Action {
		case "output":
			bufs = append(bufs, ev)
			if len(bufs) > int(bufLines) {
				bufs = bufs[1:]
			}
		case "fail":
			diag := RDFDiagnostic{
				Severity: RDFSeverityError,
			}
			for _, te := range bufs {
				diag.Message += te.Output
				if loc, ok := parseLoc(te.Output); ok {
					diag.Location = loc
				}
			}
			if diag.Location == nil {
				slog.Warn("event found but cannot determine the location", slog.String("action", ev.Action))
				continue
			}
			slog.Info("found diagnostic", slog.Any("diagnostic", diag))
			_ = enc.Encode(diag)
			diags = append(diags, diag)
			bufs = bufs[:]
		case "skip":
			diag := RDFDiagnostic{
				Severity: RDFSeverityInfo,
			}
			for _, te := range bufs {
				diag.Message += te.Output
				if loc, ok := parseLoc(te.Output); ok {
					diag.Location = loc
				}
			}
			if diag.Location == nil {
				slog.Warn("event found but cannot determine the location", slog.String("action", ev.Action))
				continue
			}
			slog.Info("found diagnostic", slog.Any("diagnostic", diag))
			_ = enc.Encode(diag)
			diags = append(diags, diag)
			bufs = bufs[:]
		case "run":
			// no-op
		default:
			bufs = bufs[:]
		}
	}
	slog.Info("found diagnostics", slog.Int("count", len(diags)))

	return 0
}

func parseLoc(s string) (*RDFLocation, bool) {
	parts := strings.SplitN(strings.TrimSpace(s), ":", 3)
	if len(parts) < 3 {
		return nil, false
	}
	loc := &RDFLocation{
		Path:  parts[0],
		Range: new(RDFRange),
	}
	var err error
	loc.Range.Start.Line, err = strconv.Atoi(parts[1])
	if err != nil {
		return nil, false
	}
	return loc, true
}

type TestEvent struct {
	Time    time.Time
	Action  string
	Package string
	Test    string
	Elapsed float64
	Output  string
}

type RDFSeverity int

const (
	RDFSeverityUnknown RDFSeverity = iota
	RDFSeverityError
	RDFSeverityWarning
	RDFSeverityInfo
)

type RDFPosition struct {
	Line   int `json:"line,omitempty"`
	Column int `json:"column,omitempty"`
}

func (p RDFPosition) String() string {
	b := new(strings.Builder)
	if p.Line > 0 {
		fmt.Fprintf(b, " line:%d", p.Line)
	}
	if p.Column > 0 {
		fmt.Fprintf(b, " column:%d", p.Column)
	}
	return b.String()
}

type RDFRange struct {
	Start RDFPosition `json:"start"`
	End   RDFPosition `json:"end"`
}

type RDFLocation struct {
	Path  string    `json:"path"`
	Range *RDFRange `json:"range"`
}

func (l *RDFLocation) String() string {
	b := new(strings.Builder)
	fmt.Fprintf(b, "path:%q", l.Path)
	if l.Range != nil {
		fmt.Fprintf(b, " start:%s end:%s", l.Range.Start, l.Range.End)
	}
	return b.String()
}

// RDFDiagnostic is an diagnostic representation of Reviewdog Diagnostic Format.
//
// refs. https://github.com/reviewdog/reviewdog/blob/master/proto/rdf/jsonschema/Diagnostic.jsonschema
type RDFDiagnostic struct {
	Message  string       `json:"message"`
	Severity RDFSeverity  `json:"severity"`
	Location *RDFLocation `json:"location"`
}

func openInputStream(i string) (io.ReadCloser, error) {
	if i == "" {
		return os.Stdin, nil
	}
	return os.Open(i)
}

func main() {
	os.Exit(run())
}
