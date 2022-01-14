/* tslint:disable */
import {
  VSphereAvailabilityZone,
} from '.';

export interface VSphereRegion {
  moid?: string;
  name?: string;
  zones?: VSphereAvailabilityZone[];
}
