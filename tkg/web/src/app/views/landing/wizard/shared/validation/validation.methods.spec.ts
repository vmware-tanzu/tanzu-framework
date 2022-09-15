/* tslint:disable: no-bitwise */

import {
    ipAddrToLong,
    isValidIp,
    isValidCidr,
    isValidFqdn,
    subnetMaskValid,
    subnetMaskMinLength,
    subnetMaskMaxLength,
    subnetMaskInRange,
    isIpInSubnet,
    cidrToRange,
    isCidrOverlapCidr,
    isCidrWithinCidr,
    ipOverlapSubnet,
    isNumericOnly,
    isValidClustername
} from "./validation.methods";

describe('Validation methods', () => {
    describe('ipAddrToLong', () => {
        it("'192.168.1.1' to 3232235777", () => {
            expect(ipAddrToLong('192.168.1.1')).toBe(3232235777);
        });
    });

    describe('isValidIp', () => {
        it("'192.168.1.1' should be valid", () => {
            expect(isValidIp('192.168.1.1')).toBeTruthy();
        });

        it("'192.168.1' should be invalid", () => {
            expect(isValidIp('192.168.1')).toBeFalsy();
        });
    });

    describe('isValidCidr', () => {
        it("'192.168.1.1/20' should be valid", () => {
            expect(isValidCidr('192.168.1.1/20')).toBeTruthy();
        });

        it("'192.168.1.1/120' should be invalid", () => {
            expect(isValidCidr('192.168.1.1/120')).toBeFalsy();
        });
        it("'192.168.1.1' should be invalid", () => {
            expect(isValidCidr('192.168.1.1')).toBeFalsy();
        });
    });

    describe('isValidFqdn', () => {
        it("'fqdn.com' should be valid", () => {
            expect(isValidFqdn('fqdn.com')).toBeTruthy();
        });

        it("'fqdn.com' should be invalid", () => {
            expect(isValidFqdn('vmware/vsphere')).toBeFalsy();
        });
    });

    describe('subnetMaskValid', () => {
        it("'255.255.255.0' should be valid", () => {
            expect(subnetMaskValid(4294967040)).toBeTruthy();
        });

        it("'192.168.1.1' should be invalid", () => {
            expect(subnetMaskValid(3232235777)).toBeFalsy();
        });
    });

    describe('subnetMaskMinLength', () => {
        it("'255.255.255.0' should satisfy min length 24", () => {
            expect(subnetMaskMinLength(4294967040, 24)).toBeTruthy();
        });

        it("'255.255.255.0' should not satisfy min length 25", () => {
            expect(subnetMaskMinLength(4294967040, 25)).toBeFalsy();
        });
    });

    describe('subnetMaskMaxLength', () => {
        it("'255.255.255.0' should satisfy max length 24", () => {
            expect(subnetMaskMaxLength(4294967040, 24)).toBeTruthy();
        });

        it("'255.255.255.0' should not satisfy max length 23", () => {
            expect(subnetMaskMaxLength(4294967040, 23)).toBeFalsy();
        });
    });

    describe('subnetMaskInRange', () => {
        it("'192.168.1.0' should be in range with '255.255.255.0'", () => {
            expect(subnetMaskInRange(3232235776, 4294967040)).toBeTruthy();
        });

        it("'192.168.1.0' should be out of range with '255.255.0.0'", () => {
            expect(subnetMaskInRange(3232235776, 4294901760)).toBeFalsy();
        });
    });

    describe('isIpInSubnet', () => {
        it("'192.168.1.1' should be in subnet of 192.168.1.0'", () => {
            expect(isIpInSubnet(3232235777, 3232235776, 4294967040)).toBeTruthy();
        });

        it("'192.168.1.1' should be out of subnet of 192.169.1.0'", () => {
            expect(isIpInSubnet(3232235777, 3232301312, 4294967040)).toBeFalsy();
        });
    });

    describe('cidrToRange', () => {
        it("'192.168.1.1/22' to [3232235520, 3232236543]", () => {
            expect(cidrToRange('192.168.1.1/22')[0] >>> 0).toEqual(3232235520);
            expect(cidrToRange('192.168.1.1/22')[1] >>> 0).toEqual(3232236543);
        });
    });

    describe('isCidrOverlapCidr', () => {
        it("'192.168.1.1/24' should overlap with '192.168.1.1/24'", () => {
            expect(isCidrOverlapCidr('10.10.1.0/1', '192.168.1.1/1')).toBeFalsy();
        });
    });

    describe('isCidrWithinCidr', () => {
        it("'192.168.1.1/20' should within '192.168.1.0/20'", () => {
            expect(isCidrWithinCidr('192.168.1.1/20', '192.168.1.0/20')).toBeTruthy();
        });

        it("'192.168.1.1/24' should not be within '192.168.2.0/24'", () => {
            expect(isCidrWithinCidr('192.168.1.1/24', '192.168.2.0/24')).toBeFalsy();
        });
    });

    describe('isNumericOnly', () => {
        it("'numeric123' should fail", () => {
            expect(isNumericOnly('numeric123')).toBeFalsy();
        });

        it("'123' should pass", () => {
            expect(isNumericOnly('123')).toBeTruthy();
        });
    });

    describe('isValidClustername', () => {
        it("'validname' should be valid", () => {
            expect(isValidClustername('validname')).toBeTruthy();
        });

        it("'validName123!!!' should not be valid", () => {
            expect(isValidClustername('validName123!!!')).toBeFalsy();
        });
    });
});
