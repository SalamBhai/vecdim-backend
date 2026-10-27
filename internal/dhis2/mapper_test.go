package dhis2_test

import (
	"testing"
	"time"

	"github.com/salambhai/vecdim-backend/ent"
	"github.com/salambhai/vecdim-backend/internal/dhis2"
	"github.com/stretchr/testify/assert"
)

func TestConvertRecordsToDHIS2(t *testing.T) {
	mort := 54.0
	who := "Resistant"
	itn := 61.5
	larval := 6
	user := "u_field_01"

	record := &ent.FieldRecord{
		ID:             "rec-001",
		UserID:         &user,
		State:          "Osun",
		Lga:            "Osogbo",
		Ward:           "Osogbo Ward 2",
		CollectionDate: "2026-10-07",
		NExposed:       100,
		NDead:          54,
		MortalityPct:   &mort,
		WhoClass:       &who,
		ItnUsePct:      &itn,
		LarvalDensity:  &larval,
		CreatedAt:      time.Date(2026, 10, 7, 9, 14, 22, 0, time.UTC),
	}

	dataValueSet := dhis2.ConvertRecordsToDHIS2([]*ent.FieldRecord{record})

	assert.NotEmpty(t, dataValueSet.DataValues)
	assert.Len(t, dataValueSet.DataValues, 6)

	elements := make(map[string]string)
	for _, dv := range dataValueSet.DataValues {
		elements[dv.DataElement] = dv.Value
		assert.Equal(t, "202610", dv.Period)
		assert.Equal(t, "Osun_Osogbo_OsogboWard2", dv.OrgUnit)
		assert.Equal(t, "u_field_01", dv.StoredBy)
	}

	assert.Equal(t, "100", elements[dhis2.DE_MosquitoExposed])
	assert.Equal(t, "54", elements[dhis2.DE_MosquitoDead])
	assert.Equal(t, "54.0", elements[dhis2.DE_MortalityPct])
	assert.Equal(t, "Resistant", elements[dhis2.DE_WHOClass])
	assert.Equal(t, "61.5", elements[dhis2.DE_ITNUsePct])
	assert.Equal(t, "6", elements[dhis2.DE_LarvalDensity])
}
