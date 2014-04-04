package model

type AttributesMixin struct {
	attributes map[string]string
}

/* Get a codec attribute value, or "" if it does not exist */
func (m *AttributesMixin) Attribute(key string) string {
	if m.attributes == nil {
		return ""
	}
	return m.attributes[key]
}

/* Returns the internal codec attributes map. */
func (m *AttributesMixin) Attributes() map[string]string {
	return m.attributes
}
