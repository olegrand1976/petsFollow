package billit

// Identifier is a Billit party identifier (Peppol / SDI / Chorus…).
type Identifier struct {
	IdentifierType string `json:"IdentifierType"`
	Identifier     string `json:"Identifier"`
}

// Address is a Billit party address.
type Address struct {
	AddressType  string `json:"AddressType"`
	Name         string `json:"Name,omitempty"`
	Street       string `json:"Street,omitempty"`
	StreetNumber string `json:"StreetNumber,omitempty"`
	City         string `json:"City,omitempty"`
	Box          string `json:"Box,omitempty"`
	CountryCode  string `json:"CountryCode,omitempty"`
	Zipcode      string `json:"Zipcode,omitempty"`
}

// CustomerDTO is the Billit customer/party embedded in an order.
type CustomerDTO struct {
	Name        string       `json:"Name"`
	VATNumber   string       `json:"VATNumber,omitempty"`
	PartyType   string       `json:"PartyType"`
	Email       string       `json:"Email,omitempty"`
	Street      string       `json:"Street,omitempty"`
	City        string       `json:"City,omitempty"`
	Zipcode     string       `json:"Zipcode,omitempty"`
	CountryCode string       `json:"CountryCode,omitempty"`
	Identifiers []Identifier `json:"Identifiers,omitempty"`
	Addresses   []Address    `json:"Addresses,omitempty"`
}

// OrderLine is a Billit order line (amounts in major currency units).
type OrderLine struct {
	Quantity      float64 `json:"Quantity"`
	UnitPriceExcl float64 `json:"UnitPriceExcl"`
	Description   string  `json:"Description"`
	VATPercentage float64 `json:"VATPercentage"`
}

// OrderDTO is the Billit POST /v1/orders body.
type OrderDTO struct {
	OrderType           string      `json:"OrderType"`
	OrderDirection      string      `json:"OrderDirection"`
	OrderNumber         string      `json:"OrderNumber,omitempty"`
	OrderDate           string      `json:"OrderDate"`
	ExpiryDate          string      `json:"ExpiryDate,omitempty"`
	Currency            string      `json:"Currency,omitempty"`
	AboutInvoiceNumber  string      `json:"AboutInvoiceNumber,omitempty"`
	Customer            CustomerDTO `json:"Customer"`
	OrderLines          []OrderLine `json:"OrderLines"`
}

// SendCommand is the body for POST /v1/orders/commands/send (shape may vary by Billit version).
type SendCommand struct {
	OrderID   int64   `json:"OrderID,omitempty"`
	OrderIDs  []int64 `json:"OrderIDs,omitempty"`
	Transport string  `json:"Transport,omitempty"`
}

// WebhookPayload is a flexible Billit status callback.
// Supports flat Order callbacks and Access Point Message/U payloads
// (EntityDetail.OrderMessage + EInvoiceFlowState).
type WebhookPayload struct {
	EventID             string `json:"EventID"`
	EventId             string `json:"eventId"`
	EventType           string `json:"EventType"`
	Event               string `json:"event"`
	Type                string `json:"type"`
	OrderID             any    `json:"OrderID"`
	OrderId             any    `json:"orderId"`
	ExternalID          string `json:"ExternalID"`
	Status              string `json:"Status"`
	PeppolStatus        string `json:"PeppolStatus"`
	DeliveryStatus      string `json:"DeliveryStatus"`
	UpdatedEntityType   string `json:"UpdatedEntityType"`
	WebhookUpdateTypeTC string `json:"WebhookUpdateTypeTC"`
	UpdatedEntityID     any    `json:"UpdatedEntityID"`
	EntityDetail        *WebhookEntityDetail `json:"EntityDetail"`
}

// WebhookEntityDetail nests Order / Message detail from Billit Access Point webhooks.
type WebhookEntityDetail struct {
	OrderMessage                 *WebhookOrderMessage `json:"OrderMessage"`
	OrderID                      any                  `json:"OrderID"`
	AdditionalMessageInformation *WebhookFlowInfo     `json:"AdditionalMessageInformation"`
	MessageAdditionalInformation *WebhookFlowInfo     `json:"MessageAdditionalInformation"`
	Description                  string               `json:"Description"`
	Success                      *bool                `json:"Success"`
}

// WebhookOrderMessage carries the Billit document id inside Message webhooks.
type WebhookOrderMessage struct {
	OrderID          any    `json:"OrderID"`
	Success          *bool  `json:"Success"`
	TransportType    string `json:"TransportType"`
	MessageDirection string `json:"MessageDirection"`
	Description      string `json:"Description"`
}

// WebhookFlowInfo holds Peppol / e-invoice network state.
type WebhookFlowInfo struct {
	EInvoiceFlowState             string `json:"EInvoiceFlowState"`
	AdditionalFlowStateInformation string `json:"AdditionalFlowStateInformation"`
}
