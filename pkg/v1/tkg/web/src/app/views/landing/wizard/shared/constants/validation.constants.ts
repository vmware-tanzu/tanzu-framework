/**
 * @enum Url pattern regex.
 * (eg: http://something.com or https://something.com)
 */
export const urlPattern = new RegExp(/^(https?:\/\/){1}([\da-z\.-]+)\.([a-z\.]{2,6})([\/\w \.-]*)*\/?$/);

/**
 * @enum IP only pattern regex.
 * (eg: 192.168.111.40)
 */
export const ipOnlyPattern = new RegExp(/^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}$/);

/**
 * @enum Host only pattern regex.
 * (eg: vsphere.local.com)
 */
export const hostOnlyPattern = new RegExp(/^(?!-)([a-zA-Z0-9-])+((\.[a-zA-Z0-9-]+)+)([a-zA-Z0-9])$/);

/**
 * @enum Host or Ip pattern regex
 * (eg: 192.168.111.40 or vsphere.local.com)
 */
// tslint:disable-next-line:max-line-length
export const hostOrIpPattern = new RegExp(/^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}$|^(?!-)([a-zA-Z0-9-])+((\.[a-zA-Z0-9-]+)+)([a-zA-Z0-9])$/);

/**
 * @enum No trailing slash pattern regex.
 * (eg: no vsphere.local.com/)
 */
export const noTrailingSlash = new RegExp(/^(.*)\/$/);

/**
 * Cluster name pattern
 */
export const clusterNamePattern = /^[a-z0-9A-Z][a-z0-9A-Z\.\-_]*[a-z0-9A-Z]$/;

/**
 * @enum ValidatorEnum error enum for different validation requirements
 */
export enum ValidatorEnum {
    // shared enums
    REQUIRED = 'required',
    WHITESPACE = 'whitespace',
    EMAIL = 'email',
    VALID_IP = 'valid IP address',
    VALID_HOST_IP = 'valid host IP address',            // Valid ip address excluding ###.###.###.0 and ###.###.###.255
    COMMA_SEPARATED_IPS = 'comma separated ips',
    COMMA_SEPARATED_IPS_OR_CIDRS = 'comma separated ips or cidrs',
    CIDR_OVERLAP_CIDR = 'cidr overlap cidr',
    CIDR_WITHIN_CIDR = 'cidr within cidr',
    VALID_FQDN = 'valid FQDN',
    VALID_IP_OR_FQDN = 'valid IP or FQDN',
    TRAILING_SLASH = 'no trailing slash',
    SUBNET_IN_RANGE = 'subnet within range',
    IP_IN_SUBNET_RANGE = 'IP in subnet range',
    IP_NOT_IN_SUBNET_RANGE = 'IP not in subnet range',
    VALID_PORT = 'valid port',
    IP_RANGE = 'ip range from small to large',
    IP_RANGE_OVERLAP = 'ip range overlap',
    CONFIRM_PASSWORD = 'confirm password',
    NUMERIC_ONLY = 'numeric only',
    PATTERN = 'pattern',
    GREATER_THAN_ZERO = 'number greater than zero',
    MIN = 'min',
    MAX = 'max',
    MIN_LEN = 'minlength',
    MAX_LEN = 'maxlength',
    NO_OVERLAP_IPS = 'no overlap ips',
    COMMA_SEPARATED_WORDS = 'comma separated words',
    INCLUDES_URL_FRAGMENT = 'no url fragment',
    INCLUDES_QUERY_PARAMS = 'no query params',
    NOT_IN_DATALIST = 'not in datalist',
    TRUE = 'true',

    // Networking enums
    NETWORKING_IP_UNIQUE = 'networking step ip unique',
    NETWORKING_NODE_IP_UNIQ = 'networking step node ip unique',
    NETWORKING_NODE_IP_SCOPE_INTERSECTION = 'networking IP range intersection',
    NETWORKING_NODE_IP_SCOPE_UNIQ = 'networking area contains network IP',
    FLOATING_IP_OVERLAP_SUBNET = 'floating ip overlap subnet',
    VLAN_OUT_OF_RANGE = 'vlan out of range',
    IP_RANGE_MIN = 'min ip count in ip range',

    // AZ step enums
    AVAILABILITY_ZONE_UNIQUE = "availability zone unique",
    VALID_RESOURCE_GROUP_NAME = "valid resource group name",
    UNIQUE_RESOURCE_GROUP_NAME = "unique resource group name",
    HTTP_OR_HTTPS = "http or https",

    // Cluster name
    VALID_CLUSTER_NAME = "cluster name valid"
}

// SimpleValidator identifies validators available from the Validation service
// which do not require special initialization (with other values or controls)
export enum SimpleValidator {
    IS_COMMA_SEPARATED_LIST,
    IS_HTTP_OR_HTTPS,
    IS_NUMBER_POSITIVE,
    IS_NUMERIC_ONLY,
    IS_STRING_WITHOUT_QUERY_PARAMS,
    IS_STRING_WITHOUT_URL_FRAGMENT,
    IS_TRUE,
    IS_VALID_CLUSTER_NAME,
    IS_VALID_FQDN,
    IS_VALID_FQDN_OR_IP,
    IS_VALID_FQDN_OR_IP_HTTPS,
    IS_VALID_FQDN_OR_IP_LIST,
    IS_VALID_FQDN_OR_IPV6,
    IS_VALID_FQDN_OR_IPV6_HTTPS,
    IS_VALID_IP,
    IS_VALID_IP_LIST,
    IS_VALID_IP_NETWORK_SEGMENT,
    IS_VALID_IPV6_NETWORK_SEGMENT,
    IS_VALID_LABEL_OR_ANNOTATION,
    IS_VALID_PORT,
    IS_VALID_RESOURCE_GROUP_NAME,
    NO_WHITE_SPACE,
    NO_TRAILING_SLASH
}
