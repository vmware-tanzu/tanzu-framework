// from https://javascript.plainenglish.io/deep-clone-an-object-and-preserve-its-type-with-typescript-d488c35e5574
export class Cloneable {
    public static deepCopy<T>(source: T): T {
        return Array.isArray(source)
            ? source.map(item => this.deepCopy(item))
            : source instanceof Date
                ? new Date(source.getTime())
                : source && typeof source === 'object'
                    ? Object.getOwnPropertyNames(source).reduce((o, prop) => {
                        Object.defineProperty(o, prop, Object.getOwnPropertyDescriptor(source, prop)!);
                        o[prop] = this.deepCopy((source as { [key: string]: any })[prop]);
                        return o;
                    }, Object.create(Object.getPrototypeOf(source)))
                    : source as T;
    }
}
