package mongodm

var DefaultLocals = map[string]string{
	"validation.field_required":           "Field '%s' is required.",
	"validation.field_invalid":            "Field '%s' has an invalid value.",
	"validation.field_invalid_id":         "Field '%s' contains an invalid object id value.",
	"validation.field_minlen":             "Field '%s' must be at least %v characters long.",
	"validation.field_maxlen":             "Field '%s' can be maximum %v characters long.",
	"validation.entry_exists":             "%s already exists for value '%v'.",
	"validation.field_not_exclusive":      "Only one of both fields can be set: '%s'' or '%s'.",
	"validation.field_required_exclusive": "Field '%s' or '%s' required.",
}
