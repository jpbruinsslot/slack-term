package slack

// SectionBlock defines a new block of type section
//
// More Information: https://api.slack.com/reference/messaging/blocks#section
type SectionBlock struct {
	Type      MessageBlockType   `json:"type"`
	Text      *TextBlockObject   `json:"text,omitempty"`
	BlockID   string             `json:"block_id,omitempty"`
	Fields    []*TextBlockObject `json:"fields,omitempty"`
	Accessory blockElement       `json:"accessory,omitempty"`
}

// blockType returns the type of the block
func (s SectionBlock) blockType() MessageBlockType {
	return s.Type
}

// NewSectionBlock returns a new instance of a section block to be rendered
func NewSectionBlock(textObj *TextBlockObject, fields []*TextBlockObject, accessory blockElement) *SectionBlock {
	return &SectionBlock{
		Type:      mbtSection,
		Text:      textObj,
		Fields:    fields,
		Accessory: accessory,
	}
}
