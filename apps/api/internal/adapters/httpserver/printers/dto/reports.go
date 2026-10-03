package dto

type ConnectorMediaCapability struct {
	ID      string `json:"id" minLength:"1" maxLength:"100"`
	Version uint32 `json:"version" minimum:"1"`
}
type ConnectorAdapterCapability struct {
	ID                 string                     `json:"id" minLength:"1" maxLength:"100"`
	ContractVersions   []uint32                   `json:"contractVersions" minItems:"1" maxItems:"16"`
	Formats            []string                   `json:"formats" minItems:"1" maxItems:"16"`
	Media              []ConnectorMediaCapability `json:"media" maxItems:"64"`
	CompletionEvidence string                     `json:"completionEvidence" minLength:"1" maxLength:"100"`
	Wake               bool                       `json:"wake"`
}
type ConnectorReport struct {
	Version      string                       `json:"version" minLength:"1" maxLength:"80"`
	Commit       string                       `json:"commit" minLength:"1" maxLength:"80"`
	Platform     string                       `json:"platform" minLength:"1" maxLength:"32"`
	Architecture string                       `json:"architecture" minLength:"1" maxLength:"32"`
	Adapters     []ConnectorAdapterCapability `json:"adapters" maxItems:"32"`
}
