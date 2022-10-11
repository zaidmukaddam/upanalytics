package renderer_test

import (
	"bytes"
	"testing"

	"upanalytics/internal/renderer"
)

func TestRenderer(t *testing.T) {
	r, err := renderer.NewRenderer(&renderer.RendererConfig{
		TemplatesFolder:  "./testdata",
		TranslationsFile: "./testdata/translations.test.yaml",
	})
	if err != nil {
		t.Errorf("%v", err)
	}

	eb := new(bytes.Buffer)
	e := "Page Title: Test Title"

	r.RenderTemplate(eb, "test", &struct{ PageTitle string }{PageTitle: "Test Title"})
	if eb.String() != e {
		t.Errorf("renderer %s != %s", eb.String(), e)
	}

	tb := new(bytes.Buffer)
	te := "test translation"

	r.RenderTemplate(tb, "translations", &struct{}{})
	if tb.String() != te {
		t.Errorf("renderer %s != %s", tb.String(), te)
	}
}
