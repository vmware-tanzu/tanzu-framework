import { Component, EventEmitter, Input, OnInit, Output } from '@angular/core';

@Component({
    selector: 'app-tree-select',
    templateUrl: './tree-select.component.html',
    styleUrls: ['./tree-select.component.scss']
})
export class TreeSelectComponent implements OnInit {
    @Input() data: any;
    @Input() selectedHandler: any;

    constructor() { }

    ngOnInit(): void {
    }

    handleClick(selected): void {
        this.selectedHandler(selected);
    }

    toggleExpand(selected): void {
        selected.isExpanded = !selected.isExpanded;
    }

}
