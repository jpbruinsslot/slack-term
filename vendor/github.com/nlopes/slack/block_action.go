package slack

// ActionBlock defines data that is used to hold interactive elements.
//
// More Information: https://api.slack.com/reference/messaging/blocks#actions
type ActionBlock struct {
	Type     MessageBlockType `json:"type"`
	BlockID  string           `json:"block_id,omitempty"`
	Elements []blockElement   `json:"elements"`
}

// blockType returns the type of the block
func (s ActionBlock) blockType() MessageBlockType {
	return s.Type
}

// NewActionBlock returns a new instance of an Action Block
func NewActionBlock(blockID string, elements ...blockElement) *ActionBlock {
	return &ActionBlock{
		Type:     mbtAction,
		BlockID:  blockID,
		Elements: elements,
	}
}
