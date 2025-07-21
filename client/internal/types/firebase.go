package types

// FirebaseConfig holds Firebase configuration for the frontend
type FirebaseConfig struct {
	ProjectID    string `json:"projectId"`
	APIKey       string `json:"apiKey"`
	AuthDomain   string `json:"authDomain"`
	EmulatorHost string `json:"emulatorHost,omitempty"`
	EmulatorUI   string `json:"emulatorUI,omitempty"`
	IsEmulator   bool   `json:"isEmulator"`
}