package dhis2

import (
	"fmt"
	"strings"

	"github.com/salambhai/vecdim-backend/ent"
	"github.com/salambhai/vecdim-backend/internal/platform/generics"
)

// DHIS2 Data Elements for NMEP National Reporting.
const (
	DE_MosquitoExposed   = "VECDIM_N_EXPOSED"
	DE_MosquitoDead      = "VECDIM_N_DEAD"
	DE_MortalityPct      = "VECDIM_MORTALITY_PCT"
	DE_WHOClass          = "VECDIM_WHO_CLASS"
	DE_ITNUsePct         = "VECDIM_ITN_USE_PCT"
	DE_LarvalDensity     = "VECDIM_LARVAL_DENSITY"
)

// DHIS2DataValue represents an atomic value in the DHIS2 dataValueSets specification.
type DHIS2DataValue struct {
	DataElement string `json:"dataElement"`
	Period      string `json:"period"` // YYYYMM
	OrgUnit     string `json:"orgUnit"`
	Value       string `json:"value"`
	StoredBy    string `json:"storedBy,omitempty"`
	Created     string `json:"created,omitempty"`
}

// DHIS2DataValueSet represents the complete payload sent to DHIS2 /api/dataValueSets.
type DHIS2DataValueSet struct {
	DataValues []DHIS2DataValue `json:"dataValues"`
}

// MapRecordToDHIS2Values transforms a single FieldRecord into standard DHIS2 data values.
func MapRecordToDHIS2Values(r *ent.FieldRecord) []DHIS2DataValue {
	// Format period as YYYYMM (e.g. 2026-10-07 -> 202610)
	period := strings.ReplaceAll(r.CollectionDate, "-", "")
	if len(period) >= 6 {
		period = period[:6]
	}

	// OrgUnit format: State_LGA_Ward
	orgUnit := fmt.Sprintf("%s_%s_%s",
		strings.ReplaceAll(r.State, " ", ""),
		strings.ReplaceAll(r.Lga, " ", ""),
		strings.ReplaceAll(r.Ward, " ", ""),
	)

	storedBy := "VECDIM_SYSTEM"
	if r.UserID != nil && *r.UserID != "" {
		storedBy = *r.UserID
	}

	created := r.CreatedAt.Format("2006-01-02T15:04:05Z")

	var values []DHIS2DataValue

	values = append(values, DHIS2DataValue{
		DataElement: DE_MosquitoExposed,
		Period:      period,
		OrgUnit:     orgUnit,
		Value:       fmt.Sprintf("%d", r.NExposed),
		StoredBy:    storedBy,
		Created:     created,
	})

	values = append(values, DHIS2DataValue{
		DataElement: DE_MosquitoDead,
		Period:      period,
		OrgUnit:     orgUnit,
		Value:       fmt.Sprintf("%d", r.NDead),
		StoredBy:    storedBy,
		Created:     created,
	})

	if r.MortalityPct != nil {
		values = append(values, DHIS2DataValue{
			DataElement: DE_MortalityPct,
			Period:      period,
			OrgUnit:     orgUnit,
			Value:       fmt.Sprintf("%.1f", *r.MortalityPct),
			StoredBy:    storedBy,
			Created:     created,
		})
	}

	if r.WhoClass != nil {
		values = append(values, DHIS2DataValue{
			DataElement: DE_WHOClass,
			Period:      period,
			OrgUnit:     orgUnit,
			Value:       *r.WhoClass,
			StoredBy:    storedBy,
			Created:     created,
		})
	}

	if r.ItnUsePct != nil {
		values = append(values, DHIS2DataValue{
			DataElement: DE_ITNUsePct,
			Period:      period,
			OrgUnit:     orgUnit,
			Value:       fmt.Sprintf("%.1f", *r.ItnUsePct),
			StoredBy:    storedBy,
			Created:     created,
		})
	}

	if r.LarvalDensity != nil {
		values = append(values, DHIS2DataValue{
			DataElement: DE_LarvalDensity,
			Period:      period,
			OrgUnit:     orgUnit,
			Value:       fmt.Sprintf("%d", *r.LarvalDensity),
			StoredBy:    storedBy,
			Created:     created,
		})
	}

	return values
}

// ConvertRecordsToDHIS2 uses the generic Map helper to convert a batch of FieldRecords into a DHIS2 DataValueSet.
func ConvertRecordsToDHIS2(records []*ent.FieldRecord) DHIS2DataValueSet {
	groupedValues := generics.Map(records, MapRecordToDHIS2Values)

	var flattened []DHIS2DataValue
	for _, group := range groupedValues {
		flattened = append(flattened, group...)
	}

	return DHIS2DataValueSet{DataValues: flattened}
}
