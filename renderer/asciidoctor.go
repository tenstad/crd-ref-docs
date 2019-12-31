package renderer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/elastic/crd-ref-docs/config"
	"github.com/elastic/crd-ref-docs/types"
)

const (
	asciidocAnchorPrefix = "{anchor_prefix}-"
)

type AsciidoctorRenderer struct {
	conf *config.Config
	*Functions
}

func NewAsciidoctorRenderer(conf *config.Config) (*AsciidoctorRenderer, error) {
	baseFuncs, err := NewFunctions(conf)
	if err != nil {
		return nil, err
	}
	return &AsciidoctorRenderer{conf: conf, Functions: baseFuncs}, nil
}

func (adr *AsciidoctorRenderer) Render(gvd []types.GroupVersionDetails) error {
	funcMap := combinedFuncMap(funcMap{prefix: "asciidoc", funcs: adr.ToFuncMap()}, funcMap{funcs: sprig.TxtFuncMap()})
	tmpl, err := loadTemplate(adr.conf.TemplatesDir, funcMap)
	if err != nil {
		return err
	}

	outputFile := adr.conf.OutputPath
	finfo, err := os.Stat(outputFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if finfo != nil && finfo.IsDir() {
		outputFile = filepath.Join(outputFile, "out.asciidoc")
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.ExecuteTemplate(f, mainTemplate, gvd)
}

func (adr *AsciidoctorRenderer) ToFuncMap() template.FuncMap {
	return template.FuncMap{
		"GroupVersionID":     adr.GroupVersionID,
		"RenderAnchorID":     adr.RenderAnchorID,
		"RenderExternalLink": adr.RenderExternalLink,
		"RenderGVLink":       adr.RenderGVLink,
		"RenderLocalLink":    adr.RenderLocalLink,
		"RenderType":         adr.RenderType,
		"RenderTypeLink":     adr.RenderTypeLink,
		"SafeID":             adr.SafeID,
		"ShouldRenderType":   adr.ShouldRenderType,
		"TypeID":             adr.TypeID,
	}
}

func (adr *AsciidoctorRenderer) ShouldRenderType(t *types.Type) bool {
	return t != nil && (t.GVK != nil || len(t.References) > 0)
}

func (adr *AsciidoctorRenderer) RenderType(t *types.Type) string {
	var sb strings.Builder
	switch t.Kind {
	case types.MapKind:
		sb.WriteString("object (")
		sb.WriteString("keys:")
		sb.WriteString(adr.RenderTypeLink(t.KeyType))
		sb.WriteString(", values:")
		sb.WriteString(adr.RenderTypeLink(t.ValueType))
		sb.WriteString(")")
	case types.ArrayKind, types.SliceKind:
		sb.WriteString(adr.RenderTypeLink(t.UnderlyingType))
		sb.WriteString(" array")
	default:
		sb.WriteString(adr.RenderTypeLink(t))
	}

	return sb.String()
}

func (adr *AsciidoctorRenderer) RenderTypeLink(t *types.Type) string {
	text := adr.SimplifiedTypeName(t)

	link, local := adr.LinkForType(t)
	if link == "" {
		return text
	}

	if local {
		return adr.RenderLocalLink(asciidocAnchorPrefix, link, text)
	} else {
		return adr.RenderExternalLink(link, text)
	}
}

func (adr *AsciidoctorRenderer) RenderLocalLink(prefix, link, text string) string {
	return fmt.Sprintf("xref:%s%s[$$%s$$]", prefix, link, text)
}

func (adr *AsciidoctorRenderer) RenderExternalLink(link, text string) string {
	return fmt.Sprintf("link:%s[$$%s$$]", link, text)
}

func (adr *AsciidoctorRenderer) RenderGVLink(gv types.GroupVersionDetails) string {
	return adr.RenderLocalLink(asciidocAnchorPrefix, adr.GroupVersionID(gv), gv.GroupVersionString())
}

func (adr *AsciidoctorRenderer) RenderAnchorID(id string) string {
	return fmt.Sprintf("%s%s", asciidocAnchorPrefix, adr.SafeID(id))
}
