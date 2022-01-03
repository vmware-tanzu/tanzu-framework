// Angular imports
import { Component, OnInit } from '@angular/core';
import { Validators } from '@angular/forms';
// Third party imports
import { takeUntil } from 'rxjs/operators';
import { combineLatest } from 'rxjs';
// App imports
import Broker from 'src/app/shared/service/broker';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { TkgEventType } from '../../../../shared/service/Messenger';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { VSphereDatastore } from '../../../../swagger/models/v-sphere-datastore.model';
import { VsphereField } from "../vsphere-wizard.constants";
import { VSphereFolder } from '../../../../swagger/models/v-sphere-folder.model';
import { VsphereResourceStepMapping } from './resource-step.fieldmapping';
import { VSphereWizardFormService } from 'src/app/shared/service/vsphere-wizard-form.service';

declare var sortPaths: any;

export interface ResourcePool {
    moid?: string;
    name?: string;
    checked?: boolean;
    disabled?: boolean;
    path: string;
    parentMoid: string;
    label?: string;
    resourceType: string;
    isExpanded: boolean;
    children?: Array<ResourcePool>;
  }

const DataSources = [
    TkgEventType.VSPHERE_GET_RESOURCE_POOLS,
    TkgEventType.VSPHERE_GET_COMPUTE_RESOURCE,
    TkgEventType.VSPHERE_GET_DATA_STORES,
    TkgEventType.VSPHERE_GET_VM_FOLDERS
];

const DataTargets = {
    [TkgEventType.VSPHERE_GET_RESOURCE_POOLS]: "resourcePools",
    [TkgEventType.VSPHERE_GET_COMPUTE_RESOURCE]: "computeResources",
    [TkgEventType.VSPHERE_GET_DATA_STORES]: "datastores",
    [TkgEventType.VSPHERE_GET_VM_FOLDERS]: "vmFolders"
};

enum ResourceType {
    CLUSTER = 'cluster',
    DATACENTER = 'datacenter',
    HOST = 'host',
};

@Component({
    selector: 'app-resource-step',
    templateUrl: './resource-step.component.html',
    styleUrls: ['./resource-step.component.scss']
})
export class ResourceStepComponent extends StepFormDirective implements OnInit {

    loadingResources: boolean = false;
    resourcesFetch: number = 0;
    resourcePools: Array<ResourcePool>;
    computeResources: Array<any> = [];
    datastores: Array<VSphereDatastore>;
    vmFolders: Array<VSphereFolder>;

    treeData = [];

    constructor(private wizardFormService: VSphereWizardFormService,
                private fieldMapUtilities: FieldMapUtilities,
                private validationService: ValidationService) {
        super();
    }

    ngOnInit() {
        super.ngOnInit();
        this.fieldMapUtilities.buildForm(this.formGroup, this.formName, VsphereResourceStepMapping);

        const temp = DataSources.map(source => this.wizardFormService.getErrorStream(source));
        combineLatest(...temp)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(errors => {
                this.errorNotification = errors.filter(error => error).join(" ")
            });

        /**
         * Whenever data center selection changes, reset the relevant fields
        */
        Broker.messenger.getSubject(TkgEventType.VSPHERE_DATACENTER_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(event => {
                this.resetFieldsUponDCChange();
            });

        this.initFormWithSavedData();
        DataSources.forEach(source => {
            this.wizardFormService.getDataStream(source)
                .pipe(takeUntil(this.unsubscribe))
                .subscribe(data => {
                    this[DataTargets[source]] = sortPaths(data, function (item) { return item.name; }, '/');
                    this.resourcesFetch += 1;
                    if (this.resourcesFetch === 4) {
                        this.loadingResources = false;
                    }
                    if (source === TkgEventType.VSPHERE_GET_COMPUTE_RESOURCE) {
                        this.constructResourceTree(data);
                    }
                    if (source === TkgEventType.VSPHERE_GET_VM_FOLDERS) {
                        const selectValue = data.length === 1 ? data[0].name : this.getSavedValue(VsphereField.RESOURCE_VMFOLDER, '');
                        const validators = [Validators.required,
                            this.validationService.isValidNameInList(data.map(vmFolder => vmFolder.name))];
                        this.resurrectField(VsphereField.RESOURCE_VMFOLDER, validators, selectValue);
                    }
                    if (source === TkgEventType.VSPHERE_GET_DATA_STORES) {
                        const selectValue = data.length === 1 ? data[0].name : this.getSavedValue(VsphereField.RESOURCE_DATASTORE, '');
                        const validators = [Validators.required,
                            this.validationService.isValidNameInList(data.map(vmFolder => vmFolder.name))];
                        this.resurrectField(VsphereField.RESOURCE_DATASTORE, validators, selectValue);
                    }
                });
        });
    }

    initFormWithSavedData() {
        // overwritten to avoid setting resource pool because it causes ng-valid console errors
        const resourcePoolFields: string[] = [
            VsphereField.RESOURCE_POOL,
            VsphereField.RESOURCE_DATASTORE,
            VsphereField.RESOURCE_VMFOLDER
        ];
        if (this.hasSavedData()) {
            for (const [key, control] of Object.entries(this.formGroup.controls)) {
                if (!resourcePoolFields.includes(key)) {
                    control.setValue(this.getSavedValue(key, control.value));
                }
            }
        }
    }

    loadResourceOptions() {
        this.resourcesFetch = 0;
        this.loadingResources = true;
        this.retrieveResourcePools();
        this.retrieveComputeResources();
        this.retrieveDatastores();
        this.retrieveVMFolders();
    }

    // Reset the relevant fields upon data center change
    resetFieldsUponDCChange() {
        const fieldsToReset = [VsphereField.RESOURCE_POOL.toString(), VsphereField.RESOURCE_DATASTORE, VsphereField.RESOURCE_VMFOLDER];
        // NOTE: because the saved data values MAY be applicable to the just-chosen DC,
        // we try to set the fields to the saved value
        if (this.hasSavedData()) {
            for (const [key, control] of Object.entries(this.formGroup.controls)) {
                if (fieldsToReset.includes(key)) {
                    const savedValue = this.getSavedValue(key, control.value);
                    control.setValue(savedValue);
                }
            }
        } else {
            fieldsToReset.forEach(f => this.formGroup.get(f).setValue(""));
        }
    }

    /**
     * @method retrieveResourcePools
     * helper method to refresh list of resource pools
     */
    retrieveResourcePools() {
        Broker.messenger.publish({
            type: TkgEventType.VSPHERE_GET_RESOURCE_POOLS
        });
    }

    /**
     * @method retrieveComputeResources
     * helper method to refresh list of compute resources
     */
    retrieveComputeResources() {
        Broker.messenger.publish({
            type: TkgEventType.VSPHERE_GET_COMPUTE_RESOURCE
        });
    }

    /**
     * @method retrieveDatastores
     * helper method to refresh list of datastores
     */
    retrieveDatastores() {
        Broker.messenger.publish({
            type: TkgEventType.VSPHERE_GET_DATA_STORES
        });
    }

    /**
     * @method retrieveVMFolders
     * helper method to refresh list of vm folders
     */
    retrieveVMFolders() {
        Broker.messenger.publish({
            type: TkgEventType.VSPHERE_GET_VM_FOLDERS
        });
    }

    constructResourceTree(resources: Array<ResourcePool>): void {
        const nodeMap: Map<string, Array<ResourcePool>> = new Map(); // key is parenet moid, value is a list of child nodes.
        const resourceTree: Array<ResourcePool> = [];
        resources.forEach(resource => {
            const parentMoid = resource.parentMoid;
            if (parentMoid) {
                if (nodeMap.has(parentMoid)) {
                    nodeMap.get(parentMoid).push(resource);
                } else {
                    nodeMap.set(parentMoid, [resource]);
                }
            } else {
                resourceTree.push(resource); // it contains root nodes
            }
            resource.label = resource.name;
        });
        const selectResourcePool = this.getSavedValue(VsphereField.RESOURCE_POOL, '');
        this.constructTree(resourceTree, nodeMap, selectResourcePool);
        this.treeData = this.removeDatacenter(resourceTree);
    }

    constructTree(treeNodes: Array<ResourcePool>, map: Map<string, Array<ResourcePool>>, selectResourcePool: string): void {
        if (!treeNodes || treeNodes.length <= 0) {
            return;
        }

        treeNodes.forEach(node => {
            if (node.resourceType === ResourceType.HOST || node.resourceType === ResourceType.CLUSTER) {
                node.path += '/Resources';
            }
            const childNodes = map.get(node.moid) || [];
            node.children = childNodes;
            node.isExpanded = true;
            node.checked = selectResourcePool === node.path;
            this.constructTree(childNodes, map, selectResourcePool);
        });
    }

    removeDatacenter(resourceTree: Array<ResourcePool>): Array<ResourcePool> {
        let rootNodes = [];
        resourceTree.forEach(resource => {
            if (resource.resourceType === ResourceType.DATACENTER) {
                if (resource.children.length > 0) {
                    rootNodes = [...rootNodes, ...resource.children];
                }
            } else {
                rootNodes.push(resource);
            }
        });
        return rootNodes;
    }

    handleOnClick = (selected: ResourcePool) => {
        let resourcePoolValue = '';
        if (selected.checked) {
            this.ensureOnlyOneResourceSelected(this.treeData, selected);
            resourcePoolValue = selected.path;
        }
        this.formGroup.get(VsphereField.RESOURCE_POOL).setValue(resourcePoolValue);
    }

    ensureOnlyOneResourceSelected(resourcePools: Array<ResourcePool>, selected: ResourcePool) {
        if (!resourcePools) {
            return;
        }
        resourcePools.forEach(node => {
            node.checked = node.moid === selected.moid;
            this.ensureOnlyOneResourceSelected(node.children, selected);
        });
    }

    /**
     * Get the current value of VsphereField.RESOURCE_POOL
     */
    get resourcePoolValue() {
        return this.formGroup.get(VsphereField.RESOURCE_POOL).value;
    }

    /**
     * Get the current value of VsphereField.RESOURCE_VMFOLDER
     */
    get vmFolderValue() {
        return this.formGroup.get(VsphereField.RESOURCE_VMFOLDER).value;
    }

    /**
     * Get the current value of VsphereField.RESOURCE_DATASTORE
     */
    get datastoreValue() {
        return this.formGroup.get(VsphereField.RESOURCE_DATASTORE).value;
    }

    dynamicDescription(): string {
        const vmFolder = this.getFieldValue('vmFolder', true);
        const datastore = this.getFieldValue('datastore', true);
        const resourcePool = this.getFieldValue('resourcePool', true);
        if (vmFolder && datastore && resourcePool) {
            return 'Resource Pool: ' + resourcePool + ', VM Folder: ' + vmFolder + ', Datastore: ' + datastore;
        }
        return `Specify the resources for this ${this.clusterTypeDescriptor} cluster`;
    }
}
