package pharmacy

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// VamregUsage is HUMAN | VETERINARY (ICD medicinal / foreign product).
type VamregUsage string

const (
	VamregUsageHuman      VamregUsage = "HUMAN"
	VamregUsageVeterinary VamregUsage = "VETERINARY"
)

// VamregAmrDeclarationUnit is G | ML | PIECE.
type VamregAmrDeclarationUnit string

const (
	VamregAmrUnitG     VamregAmrDeclarationUnit = "G"
	VamregAmrUnitML    VamregAmrDeclarationUnit = "ML"
	VamregAmrUnitPiece VamregAmrDeclarationUnit = "PIECE"
)

// FlexibleFloat64 accepts JSON number or numeric string (ICD examples mix both).
type FlexibleFloat64 float64

func (f *FlexibleFloat64) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		*f = 0
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			*f = 0
			return nil
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		*f = FlexibleFloat64(v)
		return nil
	}
	var v float64
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*f = FlexibleFloat64(v)
	return nil
}

func (f FlexibleFloat64) Float64() float64 { return float64(f) }

// VamregMedicinalProduct — ICD §2.1.2.2.1
type VamregMedicinalProduct struct {
	ID                           string                                      `json:"id,omitempty"`
	CTI                          string                                      `json:"cti"`
	CNK                          string                                      `json:"cnk"`
	PpnEn                        string                                      `json:"ppnEn"`
	PpnNl                        string                                      `json:"ppnNl"`
	PpnFr                        string                                      `json:"ppnFr"`
	PackSize                     string                                      `json:"packSize"`
	Usage                        VamregUsage                                 `json:"usage"`
	ActiveSubstanceStrengthUnits []VamregMedicinalProductActiveSubstanceUnit `json:"activeSubstanceStrengthUnits"`
	PharmaceuticalFormSporIds    []string                                    `json:"pharmaceuticalFormSporIds"`
	MaName                       string                                      `json:"maName"`
	MaNumber                     string                                      `json:"maNumber"`
	MahName                      string                                      `json:"mahName"`
	AmrPackUnitAmount            FlexibleFloat64                             `json:"amrPackUnitAmount"`
	AmrDeclarationUnit           VamregAmrDeclarationUnit                    `json:"amrDeclarationUnit"`
	Deprecated                   bool                                        `json:"deprecated"`
}

// VamregMedicinalProductActiveSubstanceUnit — ICD §2.1.2.2.2
type VamregMedicinalProductActiveSubstanceUnit struct {
	ActiveSubstanceSporID string `json:"activeSubstanceSporId"`
	StrengthUnitLabel     string `json:"strengthUnitLabel"`
}

// VamregMedicinalProductsTimed — ICD §2.1.2.2.3
type VamregMedicinalProductsTimed struct {
	Timestamp         time.Time                `json:"timestamp"`
	MedicinalProducts []VamregMedicinalProduct `json:"medicinalProducts"`
}

// VamregForeignMedicinalProduct — ICD §2.1.2.3.1 (non-BE)
type VamregForeignMedicinalProduct struct {
	ID                                string                             `json:"id,omitempty"`
	ForeignProductName                string                             `json:"foreignProductName"`
	ForeignMaNumber                   string                             `json:"foreignMaNumber"`
	ForeignMaHolder                   string                             `json:"foreignMaHolder"`
	PharmaceuticalFormSporID          string                             `json:"pharmaceuticalFormSporId"`
	PackSize                          string                             `json:"packSize"`
	Usage                             VamregUsage                        `json:"usage"`
	ActiveSubstanceStrengthUnitMasses []VamregForeignActiveSubstanceMass `json:"activeSubstanceStrengthUnitMasses"`
	UpdPermanentIdentifier            string                             `json:"updPermanentIdentifier"`
	UpdPackID                         string                             `json:"updPackId"`
	AmrPackUnitAmount                 FlexibleFloat64                    `json:"amrPackUnitAmount"`
	AmrDeclarationUnit                VamregAmrDeclarationUnit           `json:"amrDeclarationUnit"`
	Deprecated                        bool                               `json:"deprecated"`
}

// VamregForeignActiveSubstanceMass — ICD §2.1.2.3.2
type VamregForeignActiveSubstanceMass struct {
	ActiveSubstanceSporID string `json:"activeSubstanceSporId"`
	Strength              string `json:"strength"`
	UnitSporID            string `json:"unitSporId"`
}

// VamregForeignMedicinalProductsTimed — ICD uses foreignMedicinalProducts or nonBeMedicinalProducts.
type VamregForeignMedicinalProductsTimed struct {
	Timestamp                time.Time                       `json:"timestamp"`
	ForeignMedicinalProducts []VamregForeignMedicinalProduct `json:"foreignMedicinalProducts"`
	NonBeMedicinalProducts   []VamregForeignMedicinalProduct `json:"nonBeMedicinalProducts"`
}

// Products returns the non-empty timed array (ICD field-name variants).
func (t VamregForeignMedicinalProductsTimed) Products() []VamregForeignMedicinalProduct {
	if len(t.ForeignMedicinalProducts) > 0 {
		return t.ForeignMedicinalProducts
	}
	return t.NonBeMedicinalProducts
}

// VamregCodeLabel — ICD §2.1.2.4 coded lists (EN/NL/FR).
type VamregCodeLabel struct {
	Code       string `json:"code"`
	En         string `json:"en"`
	Nl         string `json:"nl"`
	Fr         string `json:"fr"`
	Deprecated bool   `json:"deprecated"`
}
