/**
 * A staic wrapper of localStorate APIs, whereas, the wrapper
 * hide all the internal maintenance work from the outside world.
 */
const ALL_KEYS = "TKG_KICKSTART_ALL_KEYS";
export class PersistentStore {
    /**
     * Set data for key. If the key is already existed in the store,
     * it will be overwritten silently. The method seriailizes 'value'
     * automatically using JSON.stringify(), so make sure the value is
     * serializable.
     * @param key key in the store
     * @param value value in the store
     */
    static setItem(key: string, value: any): void {
        const allKeys = PersistentStore.getItem(ALL_KEYS) || {};
        allKeys[key] = key;
        localStorage.setItem(ALL_KEYS, JSON.stringify(allKeys));
        localStorage.setItem(key, JSON.stringify(value));
    }

    /**
     * Get the value for key. The method deserializes value automatically
     * by calling JSON.parse();
     * @param key the key whose value to get
     */
    static getItem(key: string): any {
        return JSON.parse(localStorage.getItem(key));
    }

    /**
     * Remove an entry indicated by "key". There's no error even if "key"
     * is not existed in the store.
     * @param key the key to be removed from the store
     */
    static removeItem(key): void {
        const allKeys = PersistentStore.getItem(ALL_KEYS) || {};
        delete allKeys[key];
        PersistentStore.setItem(ALL_KEYS, allKeys);
        localStorage.removeItem(key);
    }

    /**
     * Removes all entries added by us.
     */
    static clear() {
        const allKeys = PersistentStore.getItem(ALL_KEYS) as Object;
        for (const key in allKeys) {
            if (allKeys.hasOwnProperty(key)) {
                localStorage.removeItem(key);
            }
        }
        localStorage.removeItem(ALL_KEYS);
    }
}
