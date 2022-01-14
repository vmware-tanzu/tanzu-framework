/* tslint:disable */
import {
  AWSRoute,
} from '.';

export interface AWSRouteTable {
  id?: string;
  routes?: AWSRoute[];
  vpcId?: string;
}
