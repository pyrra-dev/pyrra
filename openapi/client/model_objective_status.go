/*
 * Athene
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 0.0.0
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

import (
	"encoding/json"
)

// ObjectiveStatus struct for ObjectiveStatus
type ObjectiveStatus struct {
	Availability ObjectiveStatusAvailability `json:"availability"`
	Budget       ObjectiveStatusBudget       `json:"budget"`
}

// NewObjectiveStatus instantiates a new ObjectiveStatus object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewObjectiveStatus(availability ObjectiveStatusAvailability, budget ObjectiveStatusBudget) *ObjectiveStatus {
	this := ObjectiveStatus{}
	this.Availability = availability
	this.Budget = budget
	return &this
}

// NewObjectiveStatusWithDefaults instantiates a new ObjectiveStatus object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewObjectiveStatusWithDefaults() *ObjectiveStatus {
	this := ObjectiveStatus{}
	return &this
}

// GetAvailability returns the Availability field value
func (o *ObjectiveStatus) GetAvailability() ObjectiveStatusAvailability {
	if o == nil {
		var ret ObjectiveStatusAvailability
		return ret
	}

	return o.Availability
}

// GetAvailabilityOk returns a tuple with the Availability field value
// and a boolean to check if the value has been set.
func (o *ObjectiveStatus) GetAvailabilityOk() (*ObjectiveStatusAvailability, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Availability, true
}

// SetAvailability sets field value
func (o *ObjectiveStatus) SetAvailability(v ObjectiveStatusAvailability) {
	o.Availability = v
}

// GetBudget returns the Budget field value
func (o *ObjectiveStatus) GetBudget() ObjectiveStatusBudget {
	if o == nil {
		var ret ObjectiveStatusBudget
		return ret
	}

	return o.Budget
}

// GetBudgetOk returns a tuple with the Budget field value
// and a boolean to check if the value has been set.
func (o *ObjectiveStatus) GetBudgetOk() (*ObjectiveStatusBudget, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Budget, true
}

// SetBudget sets field value
func (o *ObjectiveStatus) SetBudget(v ObjectiveStatusBudget) {
	o.Budget = v
}

func (o ObjectiveStatus) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["availability"] = o.Availability
	}
	if true {
		toSerialize["budget"] = o.Budget
	}
	return json.Marshal(toSerialize)
}

type NullableObjectiveStatus struct {
	value *ObjectiveStatus
	isSet bool
}

func (v NullableObjectiveStatus) Get() *ObjectiveStatus {
	return v.value
}

func (v *NullableObjectiveStatus) Set(val *ObjectiveStatus) {
	v.value = val
	v.isSet = true
}

func (v NullableObjectiveStatus) IsSet() bool {
	return v.isSet
}

func (v *NullableObjectiveStatus) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableObjectiveStatus(val *ObjectiveStatus) *NullableObjectiveStatus {
	return &NullableObjectiveStatus{value: val, isSet: true}
}

func (v NullableObjectiveStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableObjectiveStatus) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
