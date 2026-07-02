package importplan

type SourceType string

const (
	SourceLegacyHomebox    SourceType = "legacy_homebox"
	SourceLegacyHomeboxCSV SourceType = "legacy_homebox_csv"
)

type Severity string

const (
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

type Message struct {
	Code       string
	Severity   Severity
	Summary    string
	Detail     string
	SourceID   string
	SourceName string
}

type SourceSummary struct {
	Type        SourceType
	Name        string
	BaseURL     string
	Version     string
	ImageImport string
}

type FieldDefinition struct {
	Key         string
	DisplayName string
	Type        string
}

type Asset struct {
	SourceID       string
	SourceRef      string
	Kind           string
	Title          string
	Description    string
	ParentSourceID string
	CustomFields   map[string]any
	Archived       bool
}

type Attachment struct {
	SourceID      string
	AssetSourceID string
	FileName      string
	ContentType   string
	Content       []byte
	SizeBytes     int
	Primary       bool
}

type Plan struct {
	Source      SourceSummary
	Fields      []FieldDefinition
	Assets      []Asset
	Attachments []Attachment
	Messages    []Message
}

func (p Plan) Counts() Counts {
	var counts Counts
	counts.Fields = len(p.Fields)
	for _, item := range p.Assets {
		switch item.Kind {
		case "location":
			counts.Locations++
		default:
			counts.Assets++
		}
	}
	counts.Attachments = len(p.Attachments)
	for _, message := range p.Messages {
		switch message.Severity {
		case SeverityError:
			counts.Errors++
		default:
			counts.Warnings++
		}
	}
	return counts
}

type Counts struct {
	Fields      int
	Locations   int
	Assets      int
	Attachments int
	Warnings    int
	Errors      int
}
