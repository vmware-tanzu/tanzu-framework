import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ClusterClassInfoComponent } from './cluster-class-info.component';

describe('ClusterClassInfoComponent', () => {
    let component: ClusterClassInfoComponent;
    let fixture: ComponentFixture<ClusterClassInfoComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ ClusterClassInfoComponent ]
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ClusterClassInfoComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
