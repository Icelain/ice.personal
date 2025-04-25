package markdown

import (

	"log"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"gopkg.in/yaml.v3"

)

// render markdown into html
func Render(mdstring string) string {

	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(mdstring))

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return string(markdown.Render(doc, renderer))

}

func ParseYaml(mdstring string) map[string]string {

	result := make(map[string]string)

	if err := yaml.Unmarshal([]byte(mdstring), result); err != nil {

		log.Println(err)

	}
	
	return result

}
