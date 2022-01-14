/* tslint:disable */
import {
  HTTPProxyConfiguration,
} from '.';

export interface TKGNetwork {
  clusterDNSName?: string;
  clusterNodeCIDR?: string;
  clusterPodCIDR?: string;
  clusterServiceCIDR?: string;
  cniType?: string;
  httpProxyConfiguration?: HTTPProxyConfiguration;
  networkName?: string;
}
