package html5

import (
	"bytes"
	"io"
	"reflect"

	"github.com/bytesparadise/libasciidoc/pkg/renderer"
	"github.com/bytesparadise/libasciidoc/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Render renders the given document in HTML and writes the result in the given `writer`
func Render(ctx *renderer.Context, output io.Writer) (map[string]interface{}, error) {
	return renderDocument(ctx, output)
}

func renderElement(ctx *renderer.Context, element interface{}) ([]byte, error) {
	log.Debugf("rendering element of type `%T`", element)
	switch e := element.(type) {
	case types.TableOfContentsMacro:
		return renderTableOfContent(ctx, e)
	case types.Section:
		return renderSection(ctx, e)
	case types.Preamble:
		return renderPreamble(ctx, e)
	case types.BlankLine:
		return renderBlankLine(ctx, e)
	case types.LabeledList:
		return renderLabeledList(ctx, e)
	case types.OrderedList:
		return renderOrderedList(ctx, e)
	case types.UnorderedList:
		return renderUnorderedList(ctx, e)
	case types.Paragraph:
		return renderParagraph(ctx, e)
	case types.CrossReference:
		return renderCrossReference(ctx, e)
	case types.QuotedText:
		return renderQuotedText(ctx, e)
	case types.Passthrough:
		return renderPassthrough(ctx, e)
	case types.BlockImage:
		return renderBlockImage(ctx, e)
	case types.InlineImage:
		return renderInlineImage(ctx, e)
	case types.DelimitedBlock:
		return renderDelimitedBlock(ctx, e)
	case types.LiteralBlock:
		return renderLiteralBlock(ctx, e)
	case types.InlineElements:
		return renderInlineElements(ctx, e)
	case types.Link:
		return renderLink(ctx, e)
	case types.StringElement:
		return renderStringElement(ctx, e)
	case types.DocumentAttributeDeclaration:
		// 'process' function do not return any rendered content, but may return an error
		return nil, processAttributeDeclaration(ctx, e)
	case types.DocumentAttributeReset:
		// 'process' function do not return any rendered content, but may return an error
		return nil, processAttributeReset(ctx, e)
	case types.DocumentAttributeSubstitution:
		return renderAttributeSubstitution(ctx, e)
	case types.SingleLineComment:
		return nil, nil // nothing to do
	default:
		return nil, errors.Errorf("unsupported type of element: %T", element)
	}
}

func renderPlainString(ctx *renderer.Context, element interface{}) ([]byte, error) {
	log.Debugf("rendering plain string for element of type %T", element)
	switch element := element.(type) {
	case types.SectionTitle:
		return renderPlainStringForInlineElements(ctx, element.Content)
	case types.QuotedText:
		return renderPlainStringForInlineElements(ctx, element.Elements)
	case types.InlineImage:
		return []byte(element.Macro.Alt()), nil
	case types.Link:
		return []byte(element.Text()), nil
	case types.BlankLine:
		return []byte("\n\n"), nil
	case types.StringElement:
		return []byte(element.Content), nil
	case types.Paragraph:
		return renderPlainString(ctx, element.Lines)
	case types.InlineElements:
		buff := bytes.NewBuffer(nil)
		for _, e := range element {
			plainStringElement, err := renderPlainString(ctx, e)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to render plain string for element of type %T", e)
			}
			buff.Write(plainStringElement)
		}
		return buff.Bytes(), nil
	case []types.InlineElements:
		buff := bytes.NewBuffer(nil)
		for _, e := range element {
			plainStringElement, err := renderPlainString(ctx, e)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to render plain string for element of type %T", e)
			}
			buff.Write(plainStringElement)
		}
		return buff.Bytes(), nil
	default:
		return nil, errors.Errorf("unexpectedResult type of element to process: %T", element)
	}
}

func renderPlainStringForInlineElements(ctx *renderer.Context, elements []interface{}) ([]byte, error) {
	buff := bytes.NewBuffer(nil)
	// for _, e := range discardTrailingBlankLinesInInlineElements(elements) {
	for _, e := range elements {
		plainStringElement, err := renderPlainString(ctx, e)
		if err != nil {
			return nil, errors.Wrap(err, "unable to render plain string value")
		}
		buff.Write(plainStringElement)
	}
	return buff.Bytes(), nil
}

func discardTrailingBlankLines(elements []interface{}) []interface{} {
	// discard blank lines at the end
	filteredElements := make([]interface{}, len(elements))
	copy(filteredElements, elements)
	for {
		if len(filteredElements) == 0 {
			break
		}
		if _, ok := filteredElements[len(filteredElements)-1].(types.BlankLine); ok {
			log.Debugf("element of type %T at position %d is a blank line, discarding it", len(filteredElements)-1, filteredElements[len(filteredElements)-1])
			// remove last element of the slice since it's a blankline
			filteredElements = filteredElements[:len(filteredElements)-1]
		} else {
			break
		}
	}
	return filteredElements
}

// includeNewline returns true if the given index is NOT the last entry in the given description lines, false otherwise.
// also, it ignores the element if it is a blankline, depending on the context
func includeNewline(ctx renderer.Context, index int, content interface{}) bool {
	switch reflect.TypeOf(content).Kind() {
	case reflect.Slice, reflect.Array:
		s := reflect.ValueOf(content)
		if _, match := s.Index(index).Interface().(types.BlankLine); match {
			return ctx.IncludeBlankLine()
		}
		return index < s.Len()-1
	default:
		log.Warnf("content of type %T is not an array or a slice")
		return false
	}
}

// hasID checks if the given map has an entry with key `types.AttrID`
func hasID(attributes map[string]interface{}) bool {
	_, found := attributes[types.AttrID]
	return found
}

// getID returns the value for the entry with key `types.AttrID` in the given map
func getID(attributes map[string]interface{}) string {
	id, ok := attributes[types.AttrID].(string)
	if !ok {
		return ""
	}
	return id
}
