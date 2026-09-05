package agentmodel

type CreationMode string

const (
	CreationModeRecord     CreationMode = "record"
	CreationModeAdditional CreationMode = "additional"
)

func (mode CreationMode) Effective() CreationMode {
	if mode == "" {
		return CreationModeRecord
	}
	return mode
}

func (intent Intent) validCreationIntent() bool {
	if intent.Operation != OperationCreate {
		return intent.CreationMode == "" && intent.CreationEvidence == ""
	}
	switch intent.CreationMode.Effective() {
	case CreationModeRecord:
		return intent.CreationEvidence == ""
	case CreationModeAdditional:
		return bounded(intent.CreationEvidence, 200, false)
	default:
		return false
	}
}
