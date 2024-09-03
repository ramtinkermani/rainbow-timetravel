package entity

type Record struct {
	ID   int               	`json:"id"`
	Data map[string]string 	`json:"data"`
	EffectiveDate string	`json:"effective_date"`
	CreatedDate string		`json:"created_date"`
}


func (d *Record) Copy() Record {
	values := d.Data

	newMap := map[string]string{}
	for key, value := range values {
		newMap[key] = value
	}

	return Record{
		ID:   d.ID,
		Data: newMap,
	}
}
