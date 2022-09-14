import { Injectable } from '@angular/core';

@Injectable({
    providedIn: 'root'
})
export class ContextualHelpService {

    private componentList: any = [];
    constructor() { }

    add(component: any) {
        this.componentList.push(component);
    }

    remove(id: string) {
        this.componentList = this.componentList.filter(component => component.id !== id);
    }

    open(id: string) {
        this.componentList.find(component => component.id === id).open();
    }

    close(id: string) {
        this.componentList.find(component => component.id === id).close();
    }
}
