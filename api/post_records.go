package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rainbowmga/timetravel/contexthelpers"
	"github.com/rainbowmga/timetravel/entity"
	"github.com/rainbowmga/timetravel/service"
)

// POST /records/{id}
// if the record exists, the record is updated.
// if the record doesn't exist, the record is created.
func (a *API) PostRecords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	idNumber, err := strconv.ParseInt(id, 10, 32)

	if err != nil || idNumber <= 0 {
		err := writeError(w, "invalid id; id must be a positive number", http.StatusBadRequest)
		logError(err)
		return
	}

	var body map[string]*string
	err = json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		err := writeError(w, "invalid input; could not parse json", http.StatusBadRequest)
		logError(err)
		return
	}

	// first retrieve the record
	record, err := a.records.GetRecord(
		ctx,
		int(idNumber),
	)

	if !errors.Is(err, service.ErrRecordDoesNotExist) { // record exists
		// This is an Update request (Ideally should be a PUT, but we use POST),
		// try to get the effective date for the update (if any provided) (e.g 4 months ago change happened)
		queryParams := r.URL.Query()
		effective_date := queryParams.Get("effective_date")

		// Use Context to pass this value along to the UpdateRecord()
		ctx = context.WithValue(ctx, contexthelpers.EffectiveDateKey, effective_date)

		record, err = a.records.UpdateRecord(ctx, int(idNumber), body)
	} else { // record does not exist

		// exclude the delete updates
		recordMap := map[string]string{}
		for key, value := range body {
			if value != nil {
				recordMap[key] = *value
			}
		}

		record = entity.Record{
			ID:   int(idNumber),
			Data: recordMap,
		}
		err = a.records.CreateRecord(ctx, record)
	}

	if err != nil {
		errInWriting := writeError(w, ErrInternal.Error(), http.StatusInternalServerError)
		logError(err)
		logError(errInWriting)
		return
	}

	err = writeJSON(w, record, http.StatusOK)
	logError(err)
}
