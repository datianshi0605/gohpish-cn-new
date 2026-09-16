package models

import (
	"bytes"
	"net/mail"
	"net/url"
	"path"
	"strings"
	"text/template"

	qrcode "github.com/skip2/go-qrcode"
)

// TemplateContext is an interface that allows both campaigns and email
// requests to have a PhishingTemplateContext generated for them.
type TemplateContext interface {
	getFromAddress() string
	getBaseURL() string
}

// PhishingTemplateContext is the context that is sent to any template, such
// as the email or landing page content.
type PhishingTemplateContext struct {
	From        string
	URL         string
	Tracker     string
	TrackingURL string
	QRCode      string
	QRCodeURL   string
	QRCodeHTML  string
	RId         string
	BaseURL     string
	BaseRecipient
}

// NewPhishingTemplateContext returns a populated PhishingTemplateContext,
// parsing the correct fields from the provided TemplateContext and recipient.
func NewPhishingTemplateContext(ctx TemplateContext, r BaseRecipient, rid string) (PhishingTemplateContext, error) {
	f, err := mail.ParseAddress(ctx.getFromAddress())
	if err != nil {
		return PhishingTemplateContext{}, err
	}
	fn := f.Name
	if fn == "" {
		fn = f.Address
	}
	templateURL, err := ExecuteTemplate(ctx.getBaseURL(), r)
	if err != nil {
		return PhishingTemplateContext{}, err
	}

	// For the base URL, we'll reset the the path and the query
	// This will create a URL in the form of http://example.com
	baseURL, err := url.Parse(templateURL)
	if err != nil {
		return PhishingTemplateContext{}, err
	}
	baseURL.Path = ""
	baseURL.RawQuery = ""

	phishURL, _ := url.Parse(templateURL)
	q := phishURL.Query()
	q.Set(RecipientParameter, rid)
	phishURL.RawQuery = q.Encode()

	trackingURL, _ := url.Parse(templateURL)
	trackingURL.Path = path.Join(trackingURL.Path, "/track")
	trackingURL.RawQuery = q.Encode()

	qrCodeURL, _ := url.Parse(templateURL)
	qrCodeURL.Path = path.Join(qrCodeURL.Path, "/qr")
	qrCodeURL.RawQuery = q.Encode()
	qrCodeHTML, err := GenerateQRCodeHTML(phishURL.String())
	if err != nil {
		return PhishingTemplateContext{}, err
	}

	return PhishingTemplateContext{
		BaseRecipient: r,
		BaseURL:       baseURL.String(),
		URL:           phishURL.String(),
		TrackingURL:   trackingURL.String(),
		Tracker:       "<img alt='' style='display: none' src='" + trackingURL.String() + "'/>",
		QRCodeURL:     qrCodeURL.String(),
		QRCode:        "<img alt='信息核验二维码' width='220' height='220' src='" + qrCodeURL.String() + "' style='display:block;width:220px;height:220px;margin:0 auto;border:1px solid #e5e5e5;padding:8px;background:#fff;'/>",
		QRCodeHTML:    qrCodeHTML,
		From:          fn,
		RId:           rid,
	}, nil
}

func GenerateQRCodeHTML(content string) (string, error) {
	code, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return "", err
	}

	bitmap := code.Bitmap()
	var builder strings.Builder
	builder.WriteString(`<table role="presentation" cellpadding="0" cellspacing="0" border="0" style="border-collapse:collapse;border-spacing:0;margin:0 auto;background:#ffffff;">`)
	for _, row := range bitmap {
		builder.WriteString(`<tr>`)
		for _, dark := range row {
			color := "#ffffff"
			if dark {
				color = "#000000"
			}
			builder.WriteString(`<td style="width:5px;height:5px;line-height:5px;font-size:0;background:` + color + `;">&nbsp;</td>`)
		}
		builder.WriteString(`</tr>`)
	}
	builder.WriteString(`</table>`)
	return builder.String(), nil
}

// ExecuteTemplate creates a templated string based on the provided
// template body and data.
func ExecuteTemplate(text string, data interface{}) (string, error) {
	buff := bytes.Buffer{}
	tmpl, err := template.New("template").Parse(text)
	if err != nil {
		return buff.String(), err
	}
	err = tmpl.Execute(&buff, data)
	return buff.String(), err
}

// ValidationContext is used for validating templates and pages
type ValidationContext struct {
	FromAddress string
	BaseURL     string
}

func (vc ValidationContext) getFromAddress() string {
	return vc.FromAddress
}

func (vc ValidationContext) getBaseURL() string {
	return vc.BaseURL
}

// ValidateTemplate ensures that the provided text in the page or template
// uses the supported template variables correctly.
func ValidateTemplate(text string) error {
	vc := ValidationContext{
		FromAddress: "foo@bar.com",
		BaseURL:     "http://example.com",
	}
	td := Result{
		BaseRecipient: BaseRecipient{
			Email:     "foo@bar.com",
			FirstName: "Foo",
			LastName:  "Bar",
			Position:  "Test",
		},
		RId: "123456",
	}
	ptx, err := NewPhishingTemplateContext(vc, td.BaseRecipient, td.RId)
	if err != nil {
		return err
	}
	_, err = ExecuteTemplate(text, ptx)
	if err != nil {
		return err
	}
	return nil
}
