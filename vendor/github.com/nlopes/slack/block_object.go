package slack

// Block Objects are also known as Composition Objects
//
// For more information: https://api.slack.com/reference/messaging/composition-objects

// BlockObject defines an interface that all block object types should
// implement.
// @TODO: Is this interface needed?
type blockObject interface {
	validateType() MessageObjectType
}

// ImageBlockObject An element to insert an image - this element can be used
// in section and context blocks only. If you want a block with only an image
// in it, you're looking for the image block.
//
// More Information: https://api.slack.com/reference/messaging/block-elements#image
type ImageBlockObject struct {
	Type     MessageObjectType `json:"type"`
	ImageURL string            `json:"image_url"`
	AltText  string            `json:"alt_text"`
}

// validateType enforces block objects for element and block parameters
func (s ImageBlockObject) validateType() MessageObjectType {
	return s.Type
}

// NewImageBlockObject returns a new instance of an image block element
func NewImageBlockObject(imageURL, altText string) *ImageBlockObject {
	return &ImageBlockObject{
		Type:     motImage,
		ImageURL: imageURL,
		AltText:  altText,
	}
}

// TextBlockObject defines a text element object to be used with blocks
//
// More Information: https://api.slack.com/reference/messaging/composition-objects#text
type TextBlockObject struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	Emoji    bool   `json:"emoji,omitempty"`
	Verbatim bool   `json:"verbatim,omitempty"`
}

// validateType enforces block objects for element and block parameters
func (s TextBlockObject) validateType() MessageObjectType {
	return MessageObjectType(s.Type)
}

// NewTextBlockObject returns an instance of a new Text Block Object
func NewTextBlockObject(elementType, text string, emoji, verbatim bool) *TextBlockObject {
	return &TextBlockObject{
		Type:     elementType,
		Text:     text,
		Emoji:    emoji,
		Verbatim: verbatim,
	}
}

// ConfirmationBlockObject defines a dialog that provides a confirmation step to
// any interactive element. This dialog will ask the user to confirm their action by
// offering a confirm and deny buttons.
//
// More Information: https://api.slack.com/reference/messaging/composition-objects#confirm
type ConfirmationBlockObject struct {
	Title   *TextBlockObject `json:"title"`
	Text    *TextBlockObject `json:"text"`
	Confirm *TextBlockObject `json:"confirm"`
	Deny    *TextBlockObject `json:"deny"`
}

// validateType enforces block objects for element and block parameters
func (s ConfirmationBlockObject) validateType() MessageObjectType {
	return motConfirmation
}

// NewConfirmationBlockObject returns an instance of a new Confirmation Block Object
func NewConfirmationBlockObject(title, text, confirm, deny *TextBlockObject) *ConfirmationBlockObject {
	return &ConfirmationBlockObject{
		Title:   title,
		Text:    text,
		Confirm: confirm,
		Deny:    deny,
	}
}

// OptionBlockObject represents a single selectable item in a select menu
//
// More Information: https://api.slack.com/reference/messaging/composition-objects#option
type OptionBlockObject struct {
	Text  *TextBlockObject `json:"text"`
	Value string           `json:"value"`
}

// NewOptionBlockObject returns an instance of a new Option Block Element
func NewOptionBlockObject(value string, text *TextBlockObject) *OptionBlockObject {
	return &OptionBlockObject{
		Text:  text,
		Value: value,
	}
}

// validateType enforces block objects for element and block parameters
func (s OptionBlockObject) validateType() MessageObjectType {
	return motOption
}

// OptionGroupBlockObject Provides a way to group options in a select menu.
//
// More Information: https://api.slack.com/reference/messaging/composition-objects#option-group
type OptionGroupBlockObject struct {
	Label   *TextBlockObject     `json:"label"`
	Options []*OptionBlockObject `json:"options"`
}

// validateType enforces block objects for element and block parameters
func (s OptionGroupBlockObject) validateType() MessageObjectType {
	return motOptionGroup
}

// NewOptionGroupBlockElement returns an instance of a new option group block element
func NewOptionGroupBlockElement(label *TextBlockObject, options ...*OptionBlockObject) *OptionGroupBlockObject {
	return &OptionGroupBlockObject{
		Label:   label,
		Options: options,
	}
}
