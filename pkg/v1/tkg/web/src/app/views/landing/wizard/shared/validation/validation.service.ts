/**
 * Angular modules
 */
import { Injectable } from '@angular/core';
import XRegExp from 'xregexp';
import { AbstractControl } from '@angular/forms';
import { Netmask } from 'netmask';

/**
 * App imports
 */
import * as validationMethods from './validation.methods';
import { ValidatorEnum } from '../constants/validation.constants';

/**
 * @class ValidationService
 *  Base class for shared form validators
 */
@Injectable()
export class ValidationService {

    constructor() { }

    /**
     * @method isValidIp
     * NOTE: if string is empty, does not yield error
     */
    isValidIp(): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue) {
                return validationMethods.isValidIp(ctrlValue) ?
                    null : { [ValidatorEnum.VALID_IP]: true };
            }

            return null;
        }
    }

    /**
     * @method noWhitespaceOnEnds check if string has leading/trailing whitespsaces
     */
    noWhitespaceOnEnds(): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue) {
                if (ctrlValue.length !== ctrlValue.toString().trim().length) {
                    return {
                        [ValidatorEnum.WHITESPACE]: true
                    };
                }
            }

            return null;
        }
    }

    /**
     * @method noTrailingSlash check if string has trailing slash (IE https://test.net/)
     */
    noTrailingSlash(): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue) {
                if (ctrlValue.substr(ctrlValue.length - 1) === '/') {
                    return {
                        [ValidatorEnum.TRAILING_SLASH]: true
                    };
                }
            }

            return null;
        }
    }

    /**
     * @method isValidIps
     * NOTE: if string is empty, does not yield error
     */
    isValidIps(): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue) {
                const ips: Array<string> = ctrlValue.split(',');
                return ips
                    .map(ipStr => validationMethods.isValidIp(ipStr))
                    .reduce((a, b) => a && b, true) ? null : { [ValidatorEnum.VALID_IP]: true };
            }

            return null;
        }
    }

    /**
     * @method isValidFqdn
     * NOTE: if string is empty, does not yield error
     */
    isValidFqdn(): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue) {
                return validationMethods.isValidFqdn(ctrlValue) ?
                    null : { [ValidatorEnum.VALID_FQDN]: true };
            }

            return null;
        }
    }

    /**
     * @method isValidIpOrFqdn validator to check if input is valid IP or FQDN
     */
    isValidIpOrFqdn(): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue) {
                if (validationMethods.isValidIp(ctrlValue) ||
                    validationMethods.isValidFqdn(ctrlValue)) {
                    return null;
                }

                return {
                    [ValidatorEnum.VALID_IP_OR_FQDN]: true
                };
            }
        }
    }

    /**
     * @method isValidIpOrFqdnWithProtocol validator to check if input is valid IP or FQDN
     * with protocol prefix
     */
    isValidIpOrFqdnWithHttpsProtocol(): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue) {
                if (validationMethods.isValidIpWithHttpsProtocol(ctrlValue) ||
                    validationMethods.isValidFqdnWithHttpsProtocol(ctrlValue)) {
                    return null;
                }

                return {
                    [ValidatorEnum.VALID_IP_OR_FQDN]: true
                };
            }
        }
    }

    /**
     * @method isStringWithoutUrlFragment validator to check if input includes URL fragment
     */
    isStringWithoutUrlFragment(): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue) {
                if (validationMethods.isStringWithoutUrlFragment(ctrlValue)) {
                    return {
                        [ValidatorEnum.INCLUDES_URL_FRAGMENT]: true
                    };
                }

                return null;
            }
        }
    }

    /**
     * @method isStringWithoutQueryParams validator to check if input includes URL query params
     */
    isStringWithoutQueryParams(): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue) {
                if (validationMethods.isStringWithoutQueryParams(ctrlValue)) {
                    return {
                        [ValidatorEnum.INCLUDES_QUERY_PARAMS]: true
                    };
                }

                return null;
            }
        }
    }

    /**
     * @method isValidPort validator to check if input is valid port
     */
    isValidPort(): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue) {
                if (validationMethods.isNumericOnly(ctrlValue)) {
                    return null;
                }

                return {
                    [ValidatorEnum.VALID_PORT]: true
                };
            }

            return null;
        }
    }

    /**
     * @method isValidSubnet validator to check if input subnet is valid
     *  - valid ip
     *  - within subnet mask range
     */
    isValidSubnet(subnetMaskCtrl: AbstractControl): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue) {
                if (validationMethods.isValidIp(ctrlValue)) {
                    const ipLong = validationMethods.ipAddrToLong(ctrlValue);
                    const subnetMask = subnetMaskCtrl.value;
                    if (validationMethods.isValidIp(subnetMask)) {
                        const subnetMaskLong = validationMethods.ipAddrToLong(subnetMask);

                        if (validationMethods.subnetMaskInRange(ipLong, subnetMaskLong)) {
                            return null;
                        }

                        return {
                            [ValidatorEnum.SUBNET_IN_RANGE]: true
                        };
                    }

                    return null;
                }

                return {
                    [ValidatorEnum.VALID_IP]: true
                };
            }

            return null;
        }
    }

    /**
     * @method isValidClusterName
     * NOTE: if string start and end with a letter, and can contain only
     * lowercase letters, numbers, and hyphens.
     */
    isValidClusterName(): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue) {
                if (ctrlValue.length !== ctrlValue.toString().trim().length) {
                    return {
                        [ValidatorEnum.WHITESPACE]: true
                    };
                }
                return validationMethods.isValidClustername(ctrlValue) ?
                    null : { [ValidatorEnum.VALID_CLUSTER_NAME]: true };
            }

            return null;
        }
    }

    /**
     * @method isValidLabelOrAnnotation
     * NOTE: if string start and end with a letter, and can contain only
     * lowercase letters, numbers, hyphens, underscores, dots, and have
     * max length of 63.
     */
    isValidLabelOrAnnotation(): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue) {
                if (ctrlValue.length !== ctrlValue.toString().trim().length) {
                    return {
                        [ValidatorEnum.WHITESPACE]: true
                    };
                }
                if (ctrlValue.length > 63) {
                    return {
                        [ValidatorEnum.MAX_LEN]: true
                    }
                }
                return validationMethods.isValidLabelOrAnnotation(ctrlValue) ?
                    null : { [ValidatorEnum.VALID_CLUSTER_NAME]: true };
            }

            return null;
        }
    }

    /**
     * @method isIpInSubnet validator to check if input is within subnet range
     *  - non-empty
     *  - valid ip
     *  - within subnet range
     */
    isIpInSubnet(subnetCtrl: AbstractControl, subnetMaskCtrl: AbstractControl): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue) {
                if (validationMethods.isValidIp(ctrlValue)) {
                    const ipLong = validationMethods.ipAddrToLong(ctrlValue);
                    const subnetMask = subnetMaskCtrl.value;
                    const subnet = subnetCtrl.value;
                    if (validationMethods.isValidIp(subnet) &&
                        validationMethods.isValidIp(subnetMask)) {
                        const subnetLong = validationMethods.ipAddrToLong(subnet);
                        const subnetMaskLong = validationMethods.ipAddrToLong(subnetMask);

                        if (validationMethods.isIpInSubnet(ipLong, subnetLong, subnetMaskLong)) {
                            return null;
                        }

                        return {
                            [ValidatorEnum.IP_IN_SUBNET_RANGE]: true
                        };
                    }

                    return null;
                }

                return {
                    [ValidatorEnum.VALID_IP]: true
                };
            }

            return null;
        }
    }

    /**
     * @method isIpInSubnet2 validator to check if input is within subnet range
     * @param cidrControlName the name of the CIDR, assumming format IPv4/length, e.g: 192.167.0.0/16
     */
    isIpInSubnet2(cidrHolder: {}, cidr: string): any {
        return (control: AbstractControl) => {
            const ipv4: string = control.value;
            if (ipv4) {
                if (validationMethods.isValidIp(ipv4)) {
                    let netmask = null;
                    try {
                        netmask = new Netmask(cidrHolder[cidr]);
                    } catch (e) {
                        // the netmask may not have been initialized yet, we don't validate
                        return null;
                    }

                    if (!netmask.contains(ipv4)) {
                        return {
                            [ValidatorEnum.IP_IN_SUBNET_RANGE]: true
                        };
                    }

                    return null;
                }

                return {
                    [ValidatorEnum.VALID_IP]: true
                };
            }

            return null;
        }
    }

    /**
     * @method ipOverlapSubnet validator to check if input is overlapping with subnet
     *  - non-empty
     *  - valid ip
     *  - not overlapping with subnet
     *
     *  NOTE: non-exhaustive check
     */
    ipOverlapSubnet(subnetCtrl: AbstractControl, subnetMaskCtrl: AbstractControl): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue) {
                if (validationMethods.isValidIp(ctrlValue)) {
                    const ipLong = validationMethods.ipAddrToLong(ctrlValue);
                    const subnetMask = subnetMaskCtrl.value;
                    const subnet = subnetCtrl.value;
                    if (validationMethods.isValidIp(subnet) &&
                        validationMethods.isValidIp(subnetMask)) {
                        const subnetLong = validationMethods.ipAddrToLong(subnet);
                        const subnetMaskLong = validationMethods.ipAddrToLong(subnetMask);

                        if (validationMethods.ipOverlapSubnet(ipLong, subnetLong, subnetMaskLong)) {
                            return {
                                [ValidatorEnum.IP_NOT_IN_SUBNET_RANGE]: true
                            };
                        }
                    }

                    return null;
                }

                return {
                    [ValidatorEnum.VALID_IP]: true
                };
            }

            return null;
        }
    }

    /**
     * @method cidrOverlapCidr validator to check if input is overlapping with another cidr
     */
    cidrOverlapCidr(compareCidrCtrl: AbstractControl): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            const compareValue: string = compareCidrCtrl.value;
            if (ctrlValue) {
                if (validationMethods.isValidCidr(ctrlValue) && validationMethods.isValidCidr(compareValue)) {
                    if (validationMethods.isCidrOverlapCidr(ctrlValue, compareValue)) {
                        return {
                            [ValidatorEnum.CIDR_OVERLAP_CIDR]: true
                        };
                    }

                    return null;
                }

            }

            return null;
        }
    }

    /**
     * @method cidrWithinCidr validator to check if input is within another cidr range
     */
    cidrWithinCidr(compareCidrCtrl: AbstractControl): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            const compareValue: string = compareCidrCtrl.value;
            if (ctrlValue) {
                if (validationMethods.isValidCidr(ctrlValue) && validationMethods.isValidCidr(compareValue)) {
                    if (!validationMethods.isCidrWithinCidr(ctrlValue, compareValue)) {
                        return {
                            [ValidatorEnum.CIDR_WITHIN_CIDR]: true
                        };
                    }

                    return null;
                }

            }

            return null;
        }
    }

    /**
     * @method isValidIpNetworkSegment
     * xxx.xxx.xxx.xx/xx
     */
    isValidIpNetworkSegment(): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue) {
                const ctrlValueList = ctrlValue.split('/');
                if (ctrlValueList.length === 2) {
                    if (!validationMethods.isValidIp(ctrlValueList[0])) {
                        return {
                            [ValidatorEnum.VALID_IP]: true
                        }
                    }
                    return ctrlValueList[1] && +ctrlValueList[1] >= 0 && +ctrlValueList[1] < 32 ? null : {
                        [ValidatorEnum.VALID_IP]: true
                    }
                } else {
                    return {
                        [ValidatorEnum.VALID_IP]: true
                    };
                }
            }
            return null;
        }
    }

    /**
     * @method isIpUnique
     * @param otherControls - array of IP controls to pass in; checks all IP's to confirm they are unique
     * @returns {function(AbstractControl): {}}
     */
    isIpUnique(otherControls: Array<AbstractControl>) {
        return (control: AbstractControl) => {
            if (control.value) {

                const currentControlIp = control.value;

                for (const ipAddr of otherControls) {
                    if (currentControlIp === ipAddr.value) {
                        return { [ValidatorEnum.NETWORKING_IP_UNIQUE]: true };
                    }
                }
            }
        }
    }

    /**
     * @method confirmPassword check if password and confirm password match
     */
    confirmPassword(passwordCtrl: AbstractControl): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            const passwordValue = passwordCtrl.value;
            if (passwordCtrl) {
                if (ctrlValue) {
                    if (ctrlValue !== passwordValue) {
                        return {
                            [ValidatorEnum.CONFIRM_PASSWORD]: true
                        };
                    } else {
                        return null;
                    }
                } else {
                    return {
                        [ValidatorEnum.REQUIRED]: true
                    };
                }
            }
            return null;
        }
    }

    /**
     * @method isNumericOnly to check if input is numeric only
     */
    isNumericOnly(): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue) {
                if (validationMethods.isNumericOnly(ctrlValue)) {
                    return null;
                }

                return {
                    [ValidatorEnum.NUMERIC_ONLY]: true
                };
            }

            return null;
        }
    }

    /**
    * @method commaSeparatedIpOrFqdn
    * @param {string} arg input string
    * @return {boolean}
    */
    commaSeparatedIpOrFqdn(arg: string): any {
        const ips = arg.split(',');
        return ips.map(ip => validationMethods.isValidIp(ip) || validationMethods.isValidFqdn(ip)).reduce((a, b) => a && b, true);
    }

    /**
     * @method isCommaSeparatedIpsOrFqdn
     */
    isCommaSeparatedIpsOrFqdn(): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue) {

                if (!this.commaSeparatedIpOrFqdn(ctrlValue)) {
                    return {
                        [ValidatorEnum.VALID_IP_OR_FQDN]: true
                    };
                }
            }

            return null;
        }
    }

    /**
     * @method isNumberGreaterThanZero
     */
    isNumberGreaterThanZero(): any {
        return (control: AbstractControl) => {
            const ctrlValue: number = control.value;
            if (ctrlValue === null) {
                return null;
            }

            if (typeof ctrlValue !== 'number' || ctrlValue < 1) {
                return {
                    [ValidatorEnum.GREATER_THAN_ZERO]: true
                };
            }

            return null;
        }
    }

    isUniqueAz(otherControls: Array<AbstractControl>): any {
        return (control: AbstractControl) => {
            if (control.value) {

                const currentAz = control.value;

                for (const az of otherControls) {
                    if (currentAz === az.value) {
                        return { [ValidatorEnum.AVAILABILITY_ZONE_UNIQUE]: true };
                    }
                }
            }
        }
    }

    isValidResourceGroupName(): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue && !XRegExp('^[\\pL-_.()\\w]+$').test(ctrlValue)) {
                return { [ValidatorEnum.VALID_RESOURCE_GROUP_NAME]: true };
            }
            return null;
        }
    }

    isUniqueResourceGroupName(resourceGroups): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue && resourceGroups.find(x => x.name === ctrlValue) != null) {
                return { [ValidatorEnum.UNIQUE_RESOURCE_GROUP_NAME]: true };
            }
            return null;
        }
    }

    /**
     * @method isCommaSeperatedList
     */
    isCommaSeperatedList(): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue) {
                if (!XRegExp('^(((\\w+)(,\\s?\\w+)+)|(\\w+))$').test(ctrlValue)) {
                    return {
                        [ValidatorEnum.COMMA_SEPERATED_WORDS]: true
                    };
                }
            }
        }
    }

    isHttpOrHttps(): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (ctrlValue && !XRegExp('^https?:\/\/').test(ctrlValue)) {
                return { [ValidatorEnum.HTTP_OR_HTTPS]: true };
            }
            return null;
        }
    }

    /**
     * @method isValidLdap
     *  - non-empty
     *  - valid IP or FQDN
     *  - valid port
     *
     * @param {AbstractControl} ipCtrl
     */
    isValidLdap(ipCtrl: AbstractControl): any {
        return (control: AbstractControl) => {
            const inputVal = ipCtrl.value;
            if (inputVal) {
                if (!validationMethods.isValidIp(inputVal) && !(validationMethods.isValidFqdn(inputVal))) {
                    return {
                        [ValidatorEnum.VALID_IP_OR_FQDN]: true
                    };
                }
            } else {
                return {
                    [ValidatorEnum.REQUIRED]: true
                };
            }

            if (control.value) {
                if (!validationMethods.isNumericOnly(control.value)) {
                    return {
                        [ValidatorEnum.VALID_PORT]: true
                    };
                }
                return null;
            }

            return {
                [ValidatorEnum.REQUIRED]: true
            };
        }
    }
    isValidNameInList(list: Array<String>): any {
        return (control: AbstractControl) => {
            const ctrlValue: string = control.value;
            if (list.indexOf(ctrlValue) === -1) {
                return {
                    [ValidatorEnum.NOT_IN_DATALIST]: true
                };
            }
            return null;
        }
    }

    isTrue(): any {
        return (control: AbstractControl) => {
            const ctrlValue: boolean = control.value;
            if (ctrlValue === true) {
                 return {[ValidatorEnum.TRUE]: true};
            }
            return null;
        }
    }
}
