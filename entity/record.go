package entity

type Record struct {
	I_id int					`json:"_id"`	
	ID int               		`json:"id"`
	Data map[string]string 		`json:"data"`
	Updates map[string]string	`json:"updates"`
	EffectiveDate string		`json:"effective_date"`
	CreatedDate string			`json:"created_date"`
}


func (d *Record) Copy() Record {
	values := d.Data
	updates := d.Updates

	newDataMap := map[string]string{}
	for key, value := range values {
		newDataMap[key] = value
	}

	newUpdatesMap := map[string]string{}
	for key, value := range updates {
		newUpdatesMap[key] = value
	}

	return Record{
		I_id: d.I_id,
		ID:   d.ID,
		Data: newDataMap,
		Updates: newUpdatesMap,
		EffectiveDate: d.EffectiveDate,
		CreatedDate: d.CreatedDate,
	}
}
