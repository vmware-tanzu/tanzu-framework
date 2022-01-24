import { SimpleValidator } from '../constants/validation.constants';

export interface StepMapping {
    fieldMappings: FieldMapping[],
}

export interface BackingObjectMap {
    displayField: string,   // the field of the backing object used for the display
    valueField: string,     // the field of the backing object used for the value
    type?: string,          // the type of the backing object; if provided generates a sanity check on the backing object type
}

export interface FieldMapping {
    name: string,                   // name of field

    backingObject?: BackingObjectMap,   // data for using backing object (see notes)
    defaultValue?: any,             // default value to initialize with
    doNotAutoSave?: boolean,        // do not save this field when saving the entire mapping (field may be "manually" saved by step)
    doNotAutoRestore?: boolean,     // do not auto-restore the stored value to this field (field is usually set by change event)
    displayOnly?: boolean,          // do not auto-build/store/restore this control; value is "manually" stored and displayed to the user
    featureFlag?: string,           // a feature flag that needs to be true in order to create this field
    isBoolean?: boolean,            // does the field have a boolean value?
    isMap?: boolean,                // is the value of this field a Map<string, string>?
    label?: string,                 // label used when displaying this field, in HTML or in confirmation page (empty if not displayed)
    mask?: boolean,                 // when saving this field, should a masked value be saved instead (for password-like fields)
    neverStore?: boolean,           // used for temp fields like user input for adding key-value pairs
    primaryTrigger?: boolean,       // do NOT set value on INIT, but immediately AFTER onChange events are subscribed to
    required?: boolean,             // should a Validator.REQUIRED validator be added to the validator list?
    requiresBackendData?: boolean,  // this field requires backend data, so do not initialize but let backend data handler initialize
    validators?: SimpleValidator[], // validators used by Clarity framework
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
// backingObject:
//    If the value of this field is an OBJECT (not a string or a Map<string, string>), then backingObject specifies how to get the
//    display and value strings from that object. Note that the user needs to supply extra parameters when using methods that store or
//    restore the field's value (like buildForm(), restoreForm() or restoreField()).
//    The optional TYPE field lets the framework warn to the console if the field's value is of a different type than the mapping expects.
