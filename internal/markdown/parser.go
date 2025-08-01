package markdown

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// SegmentはMarkdownドキュメントの一部を表します。
// 翻訳対象かどうかのフラグを持ちます。
type Segment struct {
	Content        string
	IsTranslatable bool
}

// ParserはMarkdownの解析ロジックを管理します。
type Parser struct {
	gm goldmark.Markdown
}

// NewParserは新しいParserインスタンスを作成します。
func NewParser() *Parser {
	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		),
		goldmark.WithParserOptions(
			parser.WithAttribute(),
		),
	)
	return &Parser{gm: gm}
}

// ParseはMarkdownコンテンツを読み込み、翻訳可能なセグメントとそうでないセグメントに分割します。
func (p *Parser) Parse(source []byte) ([]Segment, error) {
	reader := text.NewReader(source)
	doc := p.gm.Parser().Parse(reader)

	var segments []Segment
	var lastPos int

	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		// テキストノード以外は処理しない
		if n.Kind() != ast.KindText {
			return ast.WalkContinue, nil
		}

		textNode := n.(*ast.Text)
		start := textNode.Segment.Start
		stop := textNode.Segment.Stop

		// --- SPEC CHANGE START ---
		// 親ノードをチェックして、FencedCodeBlock(```...```)の中かどうかを判断
		isTranslatable := true
		parent := n.Parent()
		if parent != nil {
			pKind := parent.Kind()
			// FencedCodeBlockの中のテキストのみを翻訳対象外とする
			// これにより、CodeSpan( `...` )の中のテキストは翻訳対象となる
			if pKind == ast.KindFencedCodeBlock {
				isTranslatable = false
			}
		}
		if textNode.IsRaw() {
			isTranslatable = false
		}
		// --- SPEC CHANGE END ---

		if start < lastPos {
			return ast.WalkContinue, nil
		}

		// 前回のノードの終わりから今回のノードの始まりまでを非翻訳セグメントとして追加
		if start > lastPos {
			segments = append(segments, Segment{
				Content:        string(source[lastPos:start]),
				IsTranslatable: false,
			})
		}

		// 今回のノードをセグメントとして追加
		content := string(source[start:stop])
		segments = append(segments, Segment{
			Content:        content,
			IsTranslatable: isTranslatable,
		})

		lastPos = stop
		return ast.WalkContinue, nil
	})

	if err != nil {
		return nil, err
	}

	// 最後のノード以降に残りの部分があれば非翻訳セグメントとして追加
	if len(source) > lastPos {
		segments = append(segments, Segment{
			Content:        string(source[lastPos:]),
			IsTranslatable: false,
		})
	}

	// Frontmatterの処理は変更なし
	if len(segments) > 0 && strings.HasPrefix(segments[0].Content, "---") {
		var frontmatter string
		endIndex := -1
		for i, seg := range segments {
			frontmatter += seg.Content
			if i > 0 && strings.Contains(seg.Content, "---") {
				endIndex = i
				break
			}
		}
		if endIndex != -1 {
			newSegments := []Segment{{Content: frontmatter, IsTranslatable: false}}
			segments = append(newSegments, segments[endIndex+1:]...)
		}
	}

	return segments, nil
}

// Reconstructはセグメントのスライスから元のMarkdownコンテンツを再構築します。
func Reconstruct(segments []Segment) string {
	var builder strings.Builder
	for _, seg := range segments {
		builder.WriteString(seg.Content)
	}
	return builder.String()
}
