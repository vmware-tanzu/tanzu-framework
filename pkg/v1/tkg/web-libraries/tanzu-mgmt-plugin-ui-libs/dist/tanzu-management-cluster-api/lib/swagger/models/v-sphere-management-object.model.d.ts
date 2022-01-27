export interface VSphereManagementObject {
    moid?: string;
    name?: string;
    parentMoid?: string;
    path?: string;
    resourceType?: 'datacenter' | 'cluster' | 'hostgroup' | 'folder' | 'respool' | 'vm' | 'datastore' | 'host' | 'network';
}
