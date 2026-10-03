package httpapi

import (
	"context"
	"net/http"

	"github.com/oapi-codegen/nullable"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// Pairing uses no human credential and exchanges one key-bound machine token.
type Pairing struct {
	server string
	client *Client
	http   *http.Client
}

func NewPairing(server string, transport *http.Client) (*Pairing, error) {
	client, err := New(server, "", transport)
	if err != nil {
		return nil, err
	}
	return &Pairing{server: server, client: client, http: transport}, nil
}
func (p *Pairing) Start(ctx context.Context, in ports.PairingRequest) (ports.PairingChallenge, error) {
	candidates := make([]generated.PairingCandidate, 0, len(in.Candidates))
	for _, v := range in.Candidates {
		candidates = append(candidates, generated.PairingCandidate{Id: v.ID, Name: v.Name, AdapterId: v.AdapterID, DeviceId: v.DeviceID})
	}
	response, err := read[generated.SuccessEnvelopePairingStarted](p.client.sdk.PostPrintConnectorPairings(ctx, generated.BeginPairingInputBody{Name: in.Name, PublicKey: in.PublicKey, Candidates: nullable.NewNullableWithValue(candidates)}))
	if err != nil {
		return ports.PairingChallenge{}, err
	}
	v := response.Data
	return ports.PairingChallenge{ID: v.Id, PollToken: v.PollToken, UserCode: v.UserCode, VerificationURL: v.VerificationUrl, ExpiresAt: v.ExpiresAt}, nil
}
func (p *Pairing) Poll(ctx context.Context, c ports.PairingChallenge) (ports.PairingState, error) {
	response, err := read[generated.SuccessEnvelopePairingStatus](p.client.sdk.GetPrintConnectorPairingsByPairingId(ctx, c.ID, &generated.GetPrintConnectorPairingsByPairingIdParams{XPairingToken: c.PollToken}))
	return ports.PairingState(response.Data.State), err
}
func (p *Pairing) Exchange(ctx context.Context, c ports.PairingChallenge, signature []byte) (ports.ConnectorRegistration, error) {
	response, err := read[generated.SuccessEnvelopePairingCredential](p.client.sdk.PostPrintConnectorPairingsByPairingIdCredential(ctx, c.ID, &generated.PostPrintConnectorPairingsByPairingIdCredentialParams{XPairingToken: c.PollToken}, generated.ExchangePairingInputBody{Signature: signature}))
	if err != nil {
		return ports.ConnectorRegistration{}, err
	}
	v := response.Data
	return ports.ConnectorRegistration{Server: p.server, TenantID: v.TenantId, InventoryID: v.InventoryId, ConnectorID: v.ConnectorId, Credential: v.Credential, ExpiresAt: v.ExpiresAt}, nil
}
func (p *Pairing) Activate(ctx context.Context, r ports.ConnectorRegistration, session string) error {
	client, err := New(p.server, r.Credential, p.http)
	if err != nil {
		return err
	}
	_, err = read[generated.SuccessEnvelopeConnector](client.sdk.PostPrintConsumerHeartbeat(ctx, nil, generated.HeartbeatInputBody{SessionId: session}))
	return err
}

var _ ports.PairingAPI = (*Pairing)(nil)
