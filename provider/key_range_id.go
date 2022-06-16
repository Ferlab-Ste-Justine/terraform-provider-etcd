package provider

import (
    "encoding/json"
	"fmt"
)

type KeyRangeId struct {
	Key             string
	RangeEnd        string
}

//Needed to absolutely ensure it is deterministic
func (id KeyRangeId ) MarshalJSON() ([]byte, error) {
    mKey, _ := json.Marshal(id.Key)
    mRangeEnd, _ := json.Marshal(id.RangeEnd)
    return []byte(fmt.Sprintf("{\"Key\":%s,\"RangeEnd\":%s}", string(mKey), string(mRangeEnd))), nil
}

func (id KeyRangeId) Serialize() string {
    out, _ := json.Marshal(id)
    return string(out)
}

func DeserializeRangeScopedStateId(id string) (KeyRangeId, error) {
    var rangeId KeyRangeId
    err := json.Unmarshal([]byte(id), &rangeId)
    return rangeId, err
}