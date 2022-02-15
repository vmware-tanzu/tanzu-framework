import { SimpleValidator } from '../constants/validation.constants';

export interface StepMapping {
    fieldMappings: FieldMapping[],
}

export interface BackingObjectMap {
    displayField: string,   // the field of the backing object used for the display
    valueField: string,     // the field of the backing object used for the value
}

export interface FieldMapping {
    name: string,                    // name of field

    backingObject?: BackingObjectMap,   // data for using backing object (see notes)
    defaultValue?: any,              // default value to initialize with
    deactivated?: boolean,           // set true if field is not used, as with worker instance type when in standalone mode
    displayOnly?: boolean,           // do not auto-build/store/restore this control; value is "manually" stored and displayed to the user
    doNotAutoRestore?: boolean,      // do not auto-restore the stored value to this field (field is usually set by change event)
    doNotAutoSave?: boolean,         // do not save this field when saving the entire mapping (field may be "manually" saved by step)
    featureFlag?: string,            // a feature flag that needs to be true in order to create this field
    hasNoDomControl?: boolean,       // this field does not populate a DOM control, so (a) it needs a retriever to get a value, and
                                     // (b) it needs a restorer to restore the value
    isBoolean?: boolean,             // does the field have a boolean value?
    isMap?: boolean,                 // is the value of this field a Map<string, string>?
    label?: string,                  // label used when displaying this field, in HTML or in confirmation page (empty if not displayed)
    mask?: boolean,                  // when saving this field, should a masked value be saved instead (for password-like fields)
    neverStore?: boolean,            // used for temp fields like user input for adding key-value pairs
    primaryTrigger?: boolean,        // do NOT set value on INIT, but immediately AFTER onChange events are subscribed to
    required?: boolean,              // should a Validator.REQUIRED validator be added to the validator list?
    requiresBackendData?: boolean,   // this field requires backend data, so do not initialize but let backend data handler initialize
    restorer?: (value: any) => void, // given a saved value (or a retrieved object) this closure will store it. Used esp w/hasNoDomControl
    retriever?: (value: any) => any, // given a saved value, this closure will retrieve a backing object.
    validators?: SimpleValidator[],  // validators used by Clarity framework
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
// restorer:
//    For fields that are stored in the COMPONENT (rather than in the DOM hierarchy), the framework needs to be supplied a closure that
//    will take the value (or retrieved object) and store it in the component (where it "lives"). Most fields store their value (or
//    backing object) in the DOM, so they have no need for a "restorer"; the field is expected to have the name "name", and is retrievable
//    from a formGroup passed in to a framework method
// retriever:
//    For fields that use a JavaScript OBJECT (not a string or a Map<string, string>), when the retrieves the saved value, it needs a way
//    to retrieve the full object using the stored value (which is a string). This "retriever" closure provides that retrieval mechanism.
