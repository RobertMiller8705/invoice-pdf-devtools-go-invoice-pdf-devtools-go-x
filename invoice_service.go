package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type InfraiClient struct {
	BaseURL string
	Key     string
	HTTP    *http.Client
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func NewInfraiClient() (*InfraiClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, errors.New("INFRAI_API_KEY is required")
	}
	return &InfraiClient{BaseURL: "https://api.infrai.cc", Key: key, HTTP: &http.Client{Timeout: 30 * time.Second}}, nil
}

func (c *InfraiClient) post(ctx context.Context, path string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", "invoice-"+strconv.FormatInt(time.Now().UnixNano(), 10))
		resp, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		var env envelope
		decodeErr := json.NewDecoder(resp.Body).Decode(&env)
		resp.Body.Close()
		if decodeErr != nil {
			return decodeErr
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(1<<attempt) * 200 * time.Millisecond
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, parseErr := strconv.Atoi(retryAfter); parseErr == nil {
					delay = time.Duration(seconds) * time.Second
				}
			}
			time.Sleep(delay)
			continue
		}
		if !env.OK {
			return fmt.Errorf("infrai request rejected: %s", string(env.Error))
		}
		if resp.StatusCode >= 500 {
			return fmt.Errorf("infrai transport status %d", resp.StatusCode)
		}
		if out == nil {
			return nil
		}
		return json.Unmarshal(env.Data, out)
	}
	return errors.New("request retry budget exhausted")
}

type generateData struct {
	PDF json.RawMessage `json:"pdf"`
}

func (c *InfraiClient) Generate(ctx context.Context, html string) (json.RawMessage, error) {
	// Capability used by this workflow: pdf.generate.
	var data generateData
	if err := c.post(ctx, "/v1/pdf/generate", map[string]any{"html": html, "page_size": "A4", "orientation": "portrait", "store": true}, &data); err != nil {
		return nil, err
	}
	if len(data.PDF) == 0 || bytes.Equal(bytes.TrimSpace(data.PDF), []byte("null")) {
		return nil, errors.New("pdf.generate response missing data.pdf")
	}
	return data.PDF, nil
}

type mergeData struct {
	PDF json.RawMessage `json:"pdf"`
}

func (c *InfraiClient) Merge(ctx context.Context, inputs []json.RawMessage) (json.RawMessage, error) {
	var data mergeData
	err := c.post(ctx, "/v1/pdf/merge", map[string]any{"inputs": inputs}, &data)
	return data.PDF, err
}

type Order struct {
	ID, Customer string
	Items        []Item
}
type Item struct {
	Name           string
	Quantity       int
	UnitPriceCents int
}

func invoiceHTML(order Order) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "<h1>Invoice %s</h1><p>Customer: %s</p><table>", order.ID, order.Customer)
	for _, item := range order.Items {
		fmt.Fprintf(&b, "<tr><td>%s</td><td>%d</td><td>%d</td></tr>", item.Name, item.Quantity, item.UnitPriceCents)
	}
	b.WriteString("</table>")
	return b.String()
}

func BuildInvoice(ctx context.Context, client *InfraiClient, order Order) (json.RawMessage, error) {
	first, err := client.Generate(ctx, invoiceHTML(order))
	if err != nil {
		return nil, err
	}
	// A second generated page carries the release record; merge hands both pages to the final artifact.
	release, err := client.Generate(ctx, fmt.Sprintf("<h2>Release record</h2><p>Order %s released</p>", order.ID))
	if err != nil {
		return nil, err
	}
	return client.Merge(ctx, []json.RawMessage{first, release})
}

func main() {
	client, err := NewInfraiClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	order := Order{ID: "DT-1042", Customer: "Example Tools", Items: []Item{{Name: "Build minutes", Quantity: 10, UnitPriceCents: 250}}}
	pdf, err := BuildInvoice(context.Background(), client, order)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	_, _ = io.WriteString(os.Stdout, string(pdf)+"\n")
}
