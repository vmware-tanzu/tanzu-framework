/* tslint:disable: max-line-length */
/* tslint:disable: no-bitwise */

/**
 * @method ipAddrToLong
 */
export const ipAddrToLong = (ipAddress: string) => {
    const ipAddressParts: Array<string> = ipAddress.split('.');
    const octetBitLength = 8;
    let mask = 0;
    for (let i = 0; i < ipAddressParts.length; i++) {
        mask = mask << octetBitLength;
        mask += parseInt(ipAddressParts[i], 10);
    }

    return (mask >>> 0);
}

/**
 * @method isValidIp decide if arg is a valid IP after trimming whitespaces
 * @return boolean
 */
export const isValidIp = (arg: string) => {
    const regexPattern =
        /^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
        return regexPattern.test(arg.trim());
}

/**
 * @method isValidIpWithHttpsProtocol decide if arg is a valid IP with https protocol prefix after trimming whitespaces
 * @return boolean
 */
export const isValidIpWithHttpsProtocol = (arg: string) => {
    const regexPattern =
        /https?:\/\/(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
    return regexPattern.test(arg.trim());
}

/**
 * @method isValidCidr decide if arg is a valid cidr
 * @return boolean
 */
export const isValidCidr = (arg: string) => {
    const argArr = arg.split('/');
    if (argArr.length === 2) {
        if (!isValidIp(argArr[0])) {
            return false;
        }
        return argArr[1] && +argArr[1] >= 0 && +argArr[1] < 32;
    } else {
        return false;
    }
}

/**
 * @method isValidFqdn decide if arg is a valid FQDN
 * @return boolean
 */
export const isValidFqdn = (arg: string) => {
    const regexPattern = /^[a-z0-9]+([-.][a-z0-9]+)*\.[a-z]{2,}$/i;
    return regexPattern.test(arg.trim());
}

/**
 * @method isValidFqdnWithHttpsProtocol decide if arg is a valid FQDN with https protocol prefix
 * @return boolean
 */
export const isValidFqdnWithHttpsProtocol = (arg: string) => {
    const regexPattern = /((https):\/)\/([a-z0-9]+([-.][a-z0-9]+)*\.[a-z]{2,})((\/\w+)*([\/].*)?)+$/i;
    return regexPattern.test(arg.trim());
}

/**
 * @method isStringWithoutFragment decide if arg includes url fragment '#somefragment'
 * @return boolean
 */
export const isStringWithoutUrlFragment = (arg: string) => {
    const regexPattern = /(#[\w\-]+)$/i;
    return regexPattern.test(arg.trim());
}

/**
 * @method isStringWithoutQueryParams decide if arg includes url query params '?param=foo'
 * @return boolean
 */
export const isStringWithoutQueryParams = (arg: string) => {
    const regexPattern = /(\?|\&)([^=]+)\=([^&]+)/g;
    return regexPattern.test(arg.trim());
}

/**
 * @method validSubnetMask check if subnet mask is consecutive 1s followed by 0s
 * @param {number} arg subnetMask in number format
 * @return {boolean} true if valid
 */
export const subnetMaskValid = (arg: number) => {
    const cidrStr = (arg >>> 0).toString(2);
    const cidrRex = /^1+0+$/;
    return cidrRex.test(cidrStr);
}

/**
 * @method subnetMaskMinLength check if subnet satisfies min bit length
 * @param {number} arg subnetMask in number format
 * @param {number} minLength
 * @return {boolean} true if valid
 */
export const subnetMaskMinLength = (arg: number, minLength: number) => {
    const length = (arg >>> 0).toString(2).indexOf('10') + 1;
    return length >= minLength;
}

/**
 * @method subnetMaskMaxLength check if subnet satisfies max bit length
 * @param {number} arg subnetMask in number format
 * @param {number} minLength
 * @return {boolean} true if valid
 */
export const subnetMaskMaxLength = (arg: number, maxLength: number) => {
    const length = (arg >>> 0).toString(2).indexOf('10') + 1;
    return length <= maxLength;
}

/**
 * @method subnetInRange check if subnet is within range of provided mask
 * @param {number} arg subnet in number format
 * @param {number} mask
 * @return {boolean} true if within range
 */
export const subnetMaskInRange = (arg: number, mask: number) => {
    return (((arg >>> 0) & (mask >>> 0)) >>> 0) === (arg >>> 0);
}

/**
 * @method isIpInSubnet check if ip is under provided subnet range
 * @param {number} arg IP in number format
 * @param {number} subnet subnet in number format
 * @param {number} subnetMask subnet mask in number format
 * @return {boolean} true if within range
 */
export const isIpInSubnet = (arg: number, subnet: number, subnetMask: number) => {
    if ((((arg >>> 0) & (subnetMask >>> 0)) >>> 0) === (subnet >>> 0)) {
        const hostId = arg ^ subnet;
        if (hostId === 0 || (((hostId | subnetMask) >>> 0).toString(2).indexOf('0') < 0)) {
            return false;
        }

        return true;
    }

    return false;
}

export const cidrToRange = (cidr: string) => {
    const cidrStrs: Array<string> = cidr.split('/');
    const ipLong = ipAddrToLong(cidrStrs[0]);
    const suffix = parseInt(cidrStrs[1], 10);
    const mask = ((-1 << (32 - suffix)));

    const start = ipLong & ((-1 << (32 - suffix)));
    const end = start + Math.pow(2, (32 - suffix)) - 1
    return [start, end];
}

/**
 * @method isCidrOverlapCidr check if cidr overlaps with another cidr
 * @param {number} arg cidr
 * @param {number} cidr comparing cidr
 * @return {boolean} true if overlaps
 */
export const isCidrOverlapCidr = (arg: string, cidr: string) => {
    const range1 = cidrToRange(arg);
    const range2 = cidrToRange(cidr);

    return !(range1[0] > range2[1] || range1[1] < range2[0]);
}

/**
 * @method isCidrWithinCidr check if cidr is within another cidr
 * @param {number} arg cidr
 * @param {number} cidr comparing cidr
 * @return {boolean} true if it is within
 */
export const isCidrWithinCidr = (arg: string, cidr: string) => {
    const range1 = cidrToRange(arg);
    const range2 = cidrToRange(cidr);

    return range1[0] >= range2[0] && range1[1] <= range2[1];
}

/**
 * @method ipOverlapSubnet check if ip overlaps with subnet
 * @param {number} arg IP in number format
 * @param {number} subnet subnet in number format
 * @param {number} subnetMask subnet mask in number format
 * @return {boolean} true if overlaps
 * NOTE: non-exhaustive check
 */
export const ipOverlapSubnet = (arg: number, subnet: number, subnetMask: number) => {
    if ((((arg >>> 0) & (subnetMask >>> 0)) >>> 0) === (subnet >>> 0)) {
        return true;
    }

    return false;
}

/**
 * @method isNumericOnly
 * @param {string} arg input string
 * @return {boolean}
 */
export const isNumericOnly = (arg: string) => {
    const regexPattern = /^[0-9]+$/;
    return regexPattern.test(arg.toString().trim());
}

/**
 * @method isValidClustername decide if arg is a valid cluster name
 * @return boolean
 */
export const isValidClustername = (arg: string) => {
    const regexPattern = /^[a-z0-9][a-z0-9-.]{0,40}[a-z0-9]$/;
    return regexPattern.test(arg.trim());
}

/**
 * @method isValidLabelOrAnnotation decide if arg is a valid cluster label
 * @return boolean
 */
export const isValidLabelOrAnnotation = (arg: string) => {
    const regexPattern = /^[a-z0-9A-Z]([a-z0-9A-Z\-\_\.]*[a-z0-9A-Z])?$/;
    return regexPattern.test(arg.trim());
}

export const isHttpsProtocol = (arg: string) => {
    return /^https:\/\//.test(arg);
}
