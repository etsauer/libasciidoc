package html5

import (
	"github.com/bytesparadise/libasciidoc/pkg/renderer"
)

// ContextualPipeline as structure that carries the renderer context along with
// the pipeline data to process in a template or in a nested template
type ContextualPipeline struct {
	Context *renderer.Context
	// The actual pipeline
	Data interface{}
}

// wrap wraps the data with the context in a new ContextualPipeline
func wrap(ctx *renderer.Context, data interface{}) *ContextualPipeline {
	return &ContextualPipeline{
		Context: ctx,
		Data:    data,
	}
}
