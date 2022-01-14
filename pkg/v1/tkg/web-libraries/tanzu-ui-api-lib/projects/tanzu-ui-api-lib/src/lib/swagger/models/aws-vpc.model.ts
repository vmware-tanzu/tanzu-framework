/* tslint:disable */
import {
  AWSNodeAz,
} from '.';

export interface AWSVpc {
  azs?: AWSNodeAz[];
  cidr?: string;
  vpcID?: string;
}
