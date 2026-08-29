package queue

import "encoding/json"

const (
	MessageTypeCompositionProcessing = "composition_processing"
)

// BaseMessage is the base structure for all queue messages
type BaseMessage struct {
	Type string `json:"type"`
}

// CompositionProcessingMessage is the message for composition processing
type CompositionProcessingMessage struct {
	BaseMessage
	ArtID         string `json:"art_id"`
	CompositionID string `json:"composition_id"`
}

func NewCompositionProcessingMessage(artID, compositionID string) *CompositionProcessingMessage {
	return &CompositionProcessingMessage{
		BaseMessage: BaseMessage{
			Type: MessageTypeCompositionProcessing,
		},
		ArtID:         artID,
		CompositionID: compositionID,
	}
}

func (m *CompositionProcessingMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

func (m *CompositionProcessingMessage) FromJSON(data []byte) error {
	return json.Unmarshal(data, m)
}
