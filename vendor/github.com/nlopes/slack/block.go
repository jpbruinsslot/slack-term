package slack

// @NOTE: Blocks are in beta and subject to change.

// More Information: https://api.slack.com/block-kit

// MessageBlockType defines a named string type to define each block type
// as a constant for use within the package.
type MessageBlockType string
type MessageElementType string
type MessageObjectType string

const (
	mbtSection MessageBlockType = "section"
	mbtDivider MessageBlockType = "divider"
	mbtImage   MessageBlockType = "image"
	mbtAction  MessageBlockType = "actions"
	mbtContext MessageBlockType = "context"

	metImage      MessageElementType = "image"
	metButton     MessageElementType = "button"
	metOverflow   MessageElementType = "overflow"
	metDatepicker MessageElementType = "datepicker"
	metSelect     MessageElementType = "static_select"

	motImage        MessageObjectType = "image"
	motConfirmation MessageObjectType = "confirmation"
	motOption       MessageObjectType = "option"
	motOptionGroup  MessageObjectType = "option_group"
)

// block defines an interface all block types should implement
// to ensure consistency between blocks.
type block interface {
	blockType() MessageBlockType
}

// NewBlockMessage creates a new Message that contains one or more blocks to be displayed
func NewBlockMessage(blocks ...block) Message {
	return Message{
		Msg: Msg{
			Blocks: blocks,
		},
	}
}

// AddBlockMessage appends a block to the end of the existing list of blocks
func AddBlockMessage(message Message, newBlk block) Message {
	message.Msg.Blocks = append(message.Msg.Blocks, newBlk)
	return message
}
