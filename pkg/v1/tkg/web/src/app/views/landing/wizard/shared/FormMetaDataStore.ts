import { PersistentStore } from './PersistentStore';

const STEP_LIST = "TKG_KICKSTART_STEP_LIST";
const FORM_LIST = "TKG_KICKSTART_FORM_LIST";
const DATA_TTL_MIN = 30;
const DATA_LAST_SAVED = "TKG_KICKSTART_DATA_LAST_SAVED_TIMESTAMP";

export interface FormMetaData {
    label: string;
    displayValue?: any;
}

export interface StepMetaData {
    title: string;
    description: string;
}

/**
 * Provide static methods to set/set step and form related
 * meta data while hide the internal details.
 */
export class FormMetaDataStore {
    /**
     * Reset form meta data for form "formName"
     * @param formName the form whose meta data to be reset
     */
    static reset(formName: string): void {
        if (formName) {
            PersistentStore.removeItem(FormMetaDataStore.getFormMetadataKey(formName));
        }
    }

    /**
     * Get the meta data for form "formName"
     * @param formName the form whose meta data to get
     */
    static getMetaData(formName: string): { [key: string]: FormMetaData } {
        return PersistentStore.getItem(FormMetaDataStore.getFormMetadataKey(formName));
    }

    /**
     * Get a specific item from a form's meta data
     * @param formName the form whose meta data to get
     * @param key the key to retrieve in the form's meta data
     */
    static getMetaDataItem(formName: string, key: string): FormMetaData {
        const formMetaData = this.getMetaData(formName);
        if (formMetaData) {
            return formMetaData[key];
        }
        return null;
    }

    /**
     * Save the meta data for form "formName"
     * @param formName the form whose meta data to save
     * @param formMetaData the meta data of the particular form
     */
    static setMetaData(formName: string, formMetaData: { [fieldName: string]: FormMetaData }): void {
        PersistentStore.setItem(FormMetaDataStore.getFormMetadataKey(formName), formMetaData);
        this.updateLastSavedTimestamp();
    }

    /**
     * Save one entry of meta data for form "formName"
     * @param formName whose meta data to be saved
     * @param fieldName the field whose meta data to be saved
     * @param formMetaData the meta data for that particular field of the form
     */
    static saveMetaDataEntry(formName: string, fieldName: string, formMetaData: FormMetaData): void {
        const metadata = FormMetaDataStore.getMetaData(formName) || {};
        metadata[fieldName] = formMetaData;
        FormMetaDataStore.setMetaData(formName, metadata);
    }

    /**
     * Delete one entry of meta data for form "formName"
     * @param formName whose meta data to be deleted
     * @param fieldName the field whose meta data to be deleted
     */
    static deleteMetaDataEntry(formName: string, fieldName: string): void {
        const metadata = FormMetaDataStore.getMetaData(formName) || {};
        delete metadata[fieldName];
        FormMetaDataStore.setMetaData(formName, metadata);
    }

    /**
     * The internal key used to save form meta data
     * @param formName the form name
     */
    private static getFormMetadataKey(formName: string): string {
        return `${formName}_metadata`;
    }

    /**
     * Get the meta data for each step as an ordered list (array)
     */
    static getStepList(): StepMetaData[] {
        return PersistentStore.getItem(STEP_LIST);
    }

    /**
     * Save the step meta data as an ordered list (array)
     * @param stepList step meta data (array)
     */
    static setStepList(stepList: StepMetaData[]) {
        PersistentStore.setItem(STEP_LIST, stepList);
    }

    /**
     * Reset the step meta data
     */
    static resetStepList() {
        PersistentStore.removeItem(STEP_LIST);
    }

    /**
     * Get all form names as ordered list (array)
     */
    static getFormList(): string[] {
        return PersistentStore.getItem(FORM_LIST);
    }

    /**
     * Update the formList with 'formName'. Append to the list if new;
     * otherwise, ignore it.
     * @param formName the form name to be added to the form list
     */
    static updateFormList(formName: string) {
        const formList = PersistentStore.getItem(FORM_LIST) || [];
        if (formList.indexOf(formName) < 0) {
            formList.push(formName);
            PersistentStore.setItem(FORM_LIST, formList);
        }
    }

    /**
     * Reset the form list
     */
    static resetFormList() {
        PersistentStore.removeItem(FORM_LIST);
    }

    /**
     * Save the latest data for form "formName"
     * @param formName the form whose data to be saved
     * @param formData the latest data for the form
     */
    static setFormData(formName: string, formData: any) {
        PersistentStore.setItem(FormMetaDataStore.getFormDataKey(formName), formData);
        this.updateLastSavedTimestamp();
    }

    /**
     * Get the data for form "formName"
     * @param formName the form whose data to get
     */
    static getFormData(formName: string): any {
        return PersistentStore.getItem(FormMetaDataStore.getFormDataKey(formName));
    }

    /**
     * Compares timestamp of last saved item to decide
     * if we should prompt the user to delete saved data
     */
    static shouldPromptClearLocalStorage(): boolean {
        if (PersistentStore.getItem(DATA_LAST_SAVED)) {
            const lastSavedDate = new Date(PersistentStore.getItem(DATA_LAST_SAVED))
            // get difference between dates in milliseconds, convert to minutes
            return ((Date.now() - lastSavedDate.getTime()) / 60000) > DATA_TTL_MIN
        }
        return false;
    }

    /**
     * Updates the timestamp marking the last time we saved an item
     */
    static updateLastSavedTimestamp() {
        PersistentStore.setItem(DATA_LAST_SAVED, Date.now())
    }

    /**
     * Deletes all saved data for each form
     */
    static deleteAllSavedData() {
        PersistentStore.clear();
        this.updateLastSavedTimestamp();
    }

    /**
     * Get the data key for form "formName"
     * @param formName form name
     */
    private static getFormDataKey(formName: string): string {
        return `${formName}_data`;
    }
}
