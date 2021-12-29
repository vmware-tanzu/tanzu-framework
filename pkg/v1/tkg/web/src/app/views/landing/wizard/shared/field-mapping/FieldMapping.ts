import { SimpleValidator } from '../constants/validation.constants';

export interface StepMapping {
    fieldMappings: FieldMapping[],
}

export interface FieldMapping {
    name: string,                   // name of field
    validators?: SimpleValidator[], // validators used by Clarity framework
    defaultValue?: any,             // default value to initialize with
    isBoolean?: boolean,            // is the field have a boolean value?
    required?: boolean,             // should a Validator.REQUIRED validator be added to the validator list?
    featureFlag?: string,           // a feature flag that needs to be true in order to create this field
    doNotAutoSave?: boolean,        // when saving the entire mapping, should this field be excluded?
    neverStore?: boolean,           // used for temp fields like adding key-value pairs
    requiresBackendData?: boolean,  // this field requires backend data, so do not initialize but let backend data handler initialize
    doNotAutoRestore?: boolean,     // do not auto-restore the stored value to this field (usually set by change event)
    primaryTrigger?: boolean,         // do NOT set value on INIT, but immediately AFTER onChange events are subscribed to
    mask?: boolean,                 // when saving this field, should a masked value be saved instead (for password-like fields)
    isMap?: boolean,                // is the value of this field a Map<string, string>?
    label?: string,                 // label used when displaying this field, in HTML or in confirmation page
}
// NOTES on FieldMapping:
// requiresBackendData:
//   This is for fields that give the user the option to pick from a list of backend resources or options. The field should use a stored
//   value only AFTER the data is retrieved from the back end (and the stored value is then in the list). Therefore, during INIT these
//   fields are not populated; the expectation is that the handler for the data-arriving-from-the-back-end event will set the field.
// doNotAutoRestore:
//   This is generally for fields that are set in value in response to an onChange event of another field. For example, if the user chooses
//   a DEV control plane flavor, then the devInstanceType will be set from the stored value. However, on INIT it makes no sense to set
//   the devInstanceType because the user has not selected that control plane flavor.
// primaryTrigger:
//   This is for fields that are intended to trigger updates to other fields. During INIT, we DON'T want to set this value immediately. We
//   want to wait until all the onChange event listeners are in place first, and THEN set this value (so it will trigger those listeners).
//   Note that this does not apply to a triggering field that ITSELF depends on other fields (or on data to arrive from the backend). These
//   dependent triggering fields should not have their values set during INIT; they should await whatever field they are dependent on to
//   send an event (or wait for the backend data to arrive). They should use doNotAutoRestore (or requiresBackendData).
