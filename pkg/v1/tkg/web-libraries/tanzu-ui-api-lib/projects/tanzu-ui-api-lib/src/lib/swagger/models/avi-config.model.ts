/* tslint:disable */
import {
  AviNetworkParams,
} from '.';

export interface AviConfig {
  ca_cert?: string;
  cloud?: string;
  controller?: string;
  controlPlaneHaProvider?: boolean;
  labels?: { [key: string]: string };
  managementClusterVipNetworkCidr?: string;
  managementClusterVipNetworkName?: string;
  network?: AviNetworkParams;
  password?: string;
  service_engine?: string;
  username?: string;
}
