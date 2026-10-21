package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestGenerateRejectsMissingPDF(t *testing.T) {
	for _, data := range []string{`{}`, `{"pdf":null}`} {
		t.Run(data, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, `{"ok":true,"data":`+data+`}`)
			}))
			defer server.Close()
			client := &InfraiClient{BaseURL: server.URL, HTTP: server.Client()}
			if _, err := client.Generate(context.Background(), "<p>Invoice</p>"); err == nil || !strings.Contains(err.Error(), "data.pdf") {
				t.Fatalf("Generate error = %v, want missing data.pdf", err)
			}
		})
	}
}

func TestMainExitsNonzeroWithoutKey(t *testing.T) {
	cmd := exec.Command("go", "run", ".")
	cmd.Env = append(os.Environ(), "INFRAI_API_KEY=")
	output, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "INFRAI_API_KEY is required") {
		t.Fatalf("output = %q, error = %v; want nonzero exit and key error", output, err)
	}
}

func TestInvoiceHTMLIncludesOrderLines(t *testing.T) {
	order := Order{ID: "DT-1042", Customer: "Data Lab", Items: []Item{{Name: "Runner", Quantity: 2, UnitPriceCents: 500}}}
	got := invoiceHTML(order)
	for _, want := range []string{"DT-1042", "Data Lab", "Runner", "<td>2</td>", "<td>500</td>"} {
		if !contains(got, want) {
			t.Fatalf("invoice HTML missing %q: %s", want, got)
		}
	}
}

func contains(s, part string) bool {
	for i := 0; i+len(part) <= len(s); i++ {
		if s[i:i+len(part)] == part {
			return true
		}
	}
	return false
}
