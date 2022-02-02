/**
 * @method filterInstanceTypes - utility method for applying a developer-defined regular expression to filter
 * instance type list.
 * @param data - original list of instance type strings to filter. If list of objects, see 'key' optional parameter below.
 * @param exp - regular expression applied to instance types for filtering.
 * @param featureFlag - boolean feature flag value. If true (feature enabled) will apply regular expression filter;
 * @param key (optional) - optional param; key that should be used to identify name of instance type per object.
 * otherwise returns original data.
 */
export const filterInstanceTypes = (data: Array<any>, exp: string, featureFlag: boolean, key?: string): any => {
    if (featureFlag === false) {
        return data;
    }
    const regex = new RegExp(exp);
    const result = data.filter((instanceType) => {
        return (key) ? !regex.test(instanceType[key]) : !regex.test(instanceType);
    });

    return result;
};
