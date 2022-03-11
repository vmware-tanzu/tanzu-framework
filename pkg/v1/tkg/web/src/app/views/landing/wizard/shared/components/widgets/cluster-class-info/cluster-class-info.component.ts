// Angular imports
import { Component, Input, OnInit } from '@angular/core';

// App imports
import clusterClassData from './cluster-class-data.json';

interface ClusterClass {
    name: string,
    categories: Array<{
        name: string
        variables: Array<{
            name: string,
            description: string
        }>
    }>
};

@Component({
    selector: 'app-cluster-class-info',
    templateUrl: './cluster-class-info.component.html',
    styleUrls: ['./cluster-class-info.component.scss']
})
export class ClusterClassInfoComponent implements OnInit {
    @Input() providerType: string;

    clusterClasses: Array<ClusterClass> = [];

    ngOnInit(): void {
        if (this.providerType) {
            this.clusterClasses = clusterClassData[this.providerType.toLowerCase()];
        }
    }
}
