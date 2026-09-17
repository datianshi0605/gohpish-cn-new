package controllers

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

func TestSendingProfilesPageRendersCompletely(t *testing.T) {
	tmpl, err := template.ParseFiles("../templates/base.html", "../templates/nav.html", "../templates/sending_profiles.html", "../templates/flashes.html")
	if err != nil {
		t.Fatal(err)
	}
	var page bytes.Buffer
	if err := tmpl.ExecuteTemplate(&page, "base", templateParams{Title: "发信配置"}); err != nil {
		t.Fatalf("sending profiles page rendering failed: %v", err)
	}
	for _, required := range []string{"您好 {{.FirstName}}", "{{.URL}}", `id="modalSubmit"`, "/js/dist/vendor.min.js", "/js/src/app/sending_profiles.js", "</html>"} {
		if !strings.Contains(page.String(), required) {
			t.Errorf("rendered page is missing %q", required)
		}
	}
}
