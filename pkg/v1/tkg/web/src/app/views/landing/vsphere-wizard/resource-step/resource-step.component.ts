/**
 * Angular Modules
 */
import { Component, OnInit, Input, Output, EventEmitter } from '@angular/core';
import {
    Validators,
    FormControl
} from '@angular/forms';
import { takeUntil } from 'rxjs/operators';
import { ClrSelectedState } from '@clr/angular'
import { combineLatest } from 'rxjs';

/**
 * App imports
 */
import { VSphereDatastore } from '../../../../swagger/models/v-sphere-datastore.model';
import { VSphereFolder } from '../../../../swagger/models/v-sphere-folder.model';
import { Messenger, TkgEventType } from '../../../../shared/service/Messenger';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { VSphereWizardFormService } from 'src/app/shared/service/vsphere-wizard-form.service';
import { ValidationService } from '../../wizard/shared/validation/validation.service';

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
    TkgEventType.GET_RESOURCE_POOLS,
    TkgEventType.GET_COMPUTE_RESOURCE,
    TkgEventType.GET_DATA_STORES,
    TkgEventType.GET_VM_FOLDERS
];

const DataTargets = {
    [TkgEventType.GET_RESOURCE_POOLS]: "resourcePools",
    [TkgEventType.GET_COMPUTE_RESOURCE]: "computeResources",
    [TkgEventType.GET_DATA_STORES]: "datastores",
    [TkgEventType.GET_VM_FOLDERS]: "vmFolders"
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

    constructor(
        private wizardFormService: VSphereWizardFormService,
        private messenger: Messenger,
        private validationService: ValidationService) {
        super();
    }

    ngOnInit() {
        super.ngOnInit();

        this.formGroup.addControl(
            'resourcePool',
            new FormControl('', [
                Validators.required
            ])
        );
        this.formGroup.addControl(
            'datastore',
            new FormControl('', [
                Validators.required
            ])
        );

        this.formGroup.addControl(
            'vmFolder',
            new FormControl('', [
                Validators.required
            ])
        );

        const temp = DataSources.map(source => this.wizardFormService.getErrorStream(source));
        combineLatest(...temp)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(errors => {
                this.errorNotification = errors.filter(error => error).join(" ")
            });

        DataSources.forEach(source => {
            this.wizardFormService.getDataStream(source)
                .pipe(takeUntil(this.unsubscribe))
                .subscribe(data => {
                    this[DataTargets[source]] = sortPaths(data, function (item) { return item.name; }, '/');
                    this.resourcesFetch += 1;
                    if (this.resourcesFetch === 4) {
                        this.loadingResources = false;
                    }
                    if (source === TkgEventType.GET_COMPUTE_RESOURCE) {
                        this.constructResourceTree(data);
                    }
                    if (source === TkgEventType.GET_VM_FOLDERS) {
                        this.resurrectField('vmFolder',
                            [Validators.required, this.validationService.isValidNameInList(data.map(vmFolder => vmFolder.name))]);
                    }
                    if (source === TkgEventType.GET_DATA_STORES) {
                        this.resurrectField('datastore',
                            [Validators.required, this.validationService.isValidNameInList(data.map(vmFolder => vmFolder.name))]);
                    }
                });
        });

        /**
         * Whenever data center selection changes, reset the relevant fields
        */
        this.messenger.getSubject(TkgEventType.DATACENTER_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(event => {
                this.resetFieldsUponDCChange();
            })
    }

    setSavedDataAfterLoad() {
        // overwritten to avoid setting resource pool because it causes ng-valid console errors
        const resetFields = [
            'resourcePool',
            'datastore',
            'vmFolder'
        ];
        if (this.hasSavedData()) {
            for (const [key, control] of Object.entries(this.formGroup.controls)) {
                if (!resetFields.includes(key)) {
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

    // Reset the relevent fields upon data center change
    resetFieldsUponDCChange() {
        const fieldsToReset = ['resourcePool', 'datastore', 'vmFolder'];
        fieldsToReset.forEach(f => this.formGroup.get(f).setValue(""));
    }

    /**
     * @method retrieveResourcePools
     * helper method to refresh list of resource pools
     */
    retrieveResourcePools() {
        this.messenger.publish({
            type: TkgEventType.GET_RESOURCE_POOLS
        });
    }

    /**
     * @method retrieveComputeResources
     * helper method to refresh list of compute resources
     */
    retrieveComputeResources() {
        this.messenger.publish({
            type: TkgEventType.GET_COMPUTE_RESOURCE
        });
    }

    /**
     * @method retrieveDatastores
     * helper method to refresh list of datastores
     */
    retrieveDatastores() {
        this.messenger.publish({
            type: TkgEventType.GET_DATA_STORES
        });
    }

    /**
     * @method retrieveVMFolders
     * helper method to refresh list of vm folders
     */
    retrieveVMFolders() {
        this.messenger.publish({
            type: TkgEventType.GET_VM_FOLDERS
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
        this.constructTree(resourceTree, nodeMap);
        this.treeData = this.removeDatacenter(resourceTree);
    }

    constructTree(treeNodes: Array<ResourcePool>, map: Map<string, Array<ResourcePool>>): void {
        if (!treeNodes || treeNodes.length <= 0) {
            return;
        }

        treeNodes.forEach(node => {

            if (node.resourceType === 'host' || node.resourceType === 'cluster') {
                node.path += '/Resources';
            }
            const childNodes = map.get(node.moid) || [];
            node.children = childNodes;
            node.isExpanded = true;
            this.constructTree(childNodes, map);
        });
    }

    removeDatacenter(resourceTree: Array<ResourcePool>): Array<ResourcePool> {
        let rootNodes = [];
        resourceTree.forEach(resource => {
            if (resource.resourceType === 'datacenter') {
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
        this.processData(this.treeData, selected);
        if (selected.checked) {
            this.formGroup.get('resourcePool').setValue(selected.path);
        } else {
            this.formGroup.get('resourcePool').setValue('');
        }
    }

    processData(data: Array<ResourcePool>, selected: ResourcePool) {
        if (!data) {
            return;
        }
        data.forEach(node => {
            if (node.moid === selected.moid && selected.checked) {
                node.checked = true;
            } else {
                node.checked = false;
            }
            this.processData(node.children, selected);
        });
    }

    /**
     * Get the current value of 'resourcePool'
     */
    get resourcePoolValue() {
        return this.formGroup.get('resourcePool').value;
    }

    /**
     * Get the current value of 'vmFolder'
     */
    get vmFolderValue() {
        return this.formGroup.get('vmFolder').value;
    }

    /**
     * Get the current value of 'datastore'
     */
    get datastoreValue() {
        return this.formGroup.get('datastore').value;
    }
}
