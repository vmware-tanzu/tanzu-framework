import { Component, Input, OnInit } from '@angular/core';
import { Key } from 'src/app/shared/constants/app.constants';

@Component({
    selector: 'app-tree-select',
    templateUrl: './tree-select.component.html',
    styleUrls: ['./tree-select.component.scss'],
})
export class TreeSelectComponent implements OnInit {
    @Input() data: any;
    @Input() selectedHandler: any;

    constructor() {}

    ngOnInit(): void {}

    handleClick(selected): void {
        this.selectedHandler(selected);
    }

    toggleExpand(selected): void {
        selected.isExpanded = !selected.isExpanded;
    }

    onKeyup(event, node): boolean {
        if (event.keyCode === Key.Enter) {
            node.checked = !node.checked;
            this.selectedHandler(node);
        } else if (event.keyCode === Key.LeftArrow) {
            node.isExpanded = true;
        } else if (event.keyCode === Key.RightArrow) {
            node.isExpanded = false;
        } else {
            return true;
        }
        return false;
    }
}
