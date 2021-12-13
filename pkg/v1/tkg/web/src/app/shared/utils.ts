
export class Utils {
    static safeString(src: string): string {
        return !(src) ? '' : src;
    }
}
